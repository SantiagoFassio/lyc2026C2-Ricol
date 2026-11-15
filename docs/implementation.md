# Implementación de Ricol

El archivo [Diseño de Ricol](design.md) describe la implementacion planeada del lenguaje antes de comenzar el desarrollo: la sintaxis propuesta para las Remote Procedure Calls, los tipos de dato nuevos, los imports, los servicios y las sesiones, y un plan de las fases del intérprete. Este documento completa [Diseño de Ricol](design.md) con la implementacion real a lo largo del proyecto.

La mayor parte del diseño se implementó como estaba planteado. Los cambios son principalmente de sintaxis (para alinear las construcciones nuevas con la sintaxis que Ricol ya tenía) y de detalle: reglas que el diseño no definía y que hubo que decidir al implementar. Cada sección resume lo que cambió y enlaza al documento que describe el tema en detalle.

## Índice

- [Cambios respecto del diseño](#cambios-respecto-del-diseño)
  - [Ejemplo introductorio](#ejemplo-introductorio)
  - [Tipos nuevos](#tipos-nuevos)
  - [Importación de módulos e interfaces](#importación-de-módulos-e-interfaces)
  - [Servicios y servidores](#servicios-y-servidores)
  - [Conexiones y sesiones](#conexiones-y-sesiones)
  - [Errores](#errores)
  - [`match`](#match)
  - [Características de la implementación del intérprete](#características-de-la-implementación-del-intérprete)
  - [Agregados que no estaban en el diseño](#agregados-que-no-estaban-en-el-diseño)
- [Diferencias con Lox](#diferencias-con-lox)
  - [Tipado estático](#tipado-estático)
  - [Sistema de importaciones](#sistema-de-importaciones)
  - [Enums y structs](#enums-y-structs)
  - [Sistema RPC](#sistema-rpc)
- [Ejemplo integrador: ta-te-ti](#ejemplo-integrador-ta-te-ti)

## Cambios respecto del diseño

Las secciones siguen el orden de [Sintaxis propuesta de manejo de Remote Procedure Calls](design.md#sintaxis-propuesta-de-manejo-de-remote-procedure-calls) en adelante.

### Ejemplo introductorio

El ejemplo de la cuenta bancaria funciona con la sintaxis implementada así:

**Interfaz** (`cuenta.ric`), sin cambios:

```ricol
struct Movimiento {
    monto: Int;
    saldoFinal: Int;
}

enum ResultadoRetiro {
    Ok(mov: Movimiento);
    SaldoInsuficiente(disponible: Int);
}

service Cuenta {
    func saldo() -> Int;
    func depositar(monto: Int) -> Movimiento;
    func retirar(monto: Int) -> ResultadoRetiro;
}
```

**Servidor** (`servidor.ric`):

```ricol
import "cuenta.ric";

server Cuenta at "0.0.0.0:3825" {
    let actual: Int = 0;

    func saldo() -> Int {
        return actual;
    }

    func depositar(monto: Int) -> Movimiento {
        actual = actual + monto;
        return Movimiento { monto: monto, saldoFinal: actual };
    }

    func retirar(monto: Int) -> ResultadoRetiro {
        if (monto > actual) {
            return ResultadoRetiro.SaldoInsuficiente(actual);
        }
        actual = actual - monto;
        return ResultadoRetiro.Ok(Movimiento { monto: -monto, saldoFinal: actual });
    }
}
```

**Cliente** (`cliente.ric`):

```ricol
import "cuenta.ric";

let conn: connection<Cuenta> = connect Cuenta at "localhost:3825";

func retirarYMostrar(c: session<Cuenta>, monto: Int) -> Bool {
    match (c.retirar(monto)) {
        Ok(mov) => {
            print "retiro ok, saldo: ";
            print mov.saldoFinal;
            print "\n";
            return True;
        }
        SaldoInsuficiente(disp) => {
            print "saldo insuficiente, disponible: ";
            print disp;
            print "\n";
            return False;
        }
    }
}

session conn timeout 2000 as c {
    c.depositar(100);
    retirarYMostrar(c, 30);
    retirarYMostrar(c, 500);
    print "saldo final: ";
    print c.saldo();
    print "\n";
} rescue e {
    match (e) {
        ConnectionFailed(msg) => print "no se pudo conectar: " + msg + "\n";
        ConnectionLost(msg) => print "se cortó la conexión, no se sabe si la última operación se hizo: " + msg + "\n";
        Timeout => print "el servidor no respondió a tiempo\n";
        ProtocolError(msg) => print "cliente y servidor no son compatibles: " + msg + "\n";
        ServerError(msg) => print "error interno del servidor: " + msg + "\n";
    }
}
```

Con el servidor levantado, el cliente imprime:

```plaintext
retiro ok, saldo: 70
saldo insuficiente, disponible: 70
saldo final: 70
```

Y sin servidor:

```plaintext
no se pudo conectar: cannot connect to Cuenta at localhost:3825: connection refused
```

Los cambios de sintaxis respecto del diseño son estos:

| Diseño | Implementación | Motivo |
| --- | --- | --- |
| `"tcp://0.0.0.0:3825"` | `"0.0.0.0:3825"` | TCP es el único transporte, así que la aclaracion es innecesaria. La dirección es un literal de la forma `host:puerto`, que se valida antes de ejecutar. |
| `session c = conn timeout 2000 {` | `session conn timeout 2000 as c {` | Se alinea con la sintaxis que Ricol ya tenía: como en `if`, `while` y `match`, la palabra reservada va seguida de la expresión sobre la que opera, y `=` se usa solo para declarar y asignar. |
| `match c.retirar(monto) {` | `match (c.retirar(monto)) {` | La expresión va entre paréntesis, igual que en `if` y `while`. |
| `print "saldo: ", c.saldo(), "\n";` | Un `print` por valor | `print` recibe una sola expresión, y `+` solo concatena strings: un `Int` se imprime con su propio `print`. |
| `if monto > actual {` | `if (monto > actual) {` | La condición de un `if` siempre va entre paréntesis. |

### Tipos nuevos

Los cuatro tipos se implementaron como estaban planteados. Las reglas que se agregaron al implementarlos son:

- **`struct`** (ver [Structs en Ricol](structs.md)):
  - Es inmutable y transportable si todos sus campos lo son, como en el diseño.
  - Se puede declarar en cualquier scope, y es visible desde la sentencia siguiente a su declaración.
  - Dos structs son el mismo tipo si vienen de la misma declaración, no por tener el mismo nombre.
  - No puede contenerse a sí mismo, ni directa ni indirectamente (ciclo).
- **`enum`** (ver [Enums en Ricol](enums.md)):
  - Las variantes pueden tener campos o no, como en el diseño.
  - A diferencia de un struct, un enum puede usarse a sí mismo en sus campos. Así se arman estructuras recursivas, como una lista: `enum Lista { Vacia; Nodo(valor: Int, resto: Lista); }`.
  - Sus campos no se acceden con `.`, solo con [`match`].
- **`connection<S>`** (ver [Conexiones](sessions.md#conexiones)):
  - Es un valor como cualquier otro: puede ser el tipo de una variable, de un parámetro, de un retorno o de un campo.
  - Dos conexiones se comparan con `==` y `!=`, por servicio y dirección. La dirección se guarda normalizada: `localhost:03825` y `localhost:3825` son la misma.
  - No es transportable, como en el diseño.
- **`session<S>`** (ver [`session<S>` no puede escapar de su bloque](sessions.md#sessions-no-puede-escapar-de-su-bloque)):
  - El diseño decía que nunca puede escapar de su bloque `session`. Para garantizarlo antes de ejecutar, `session<S>` solo puede ser el tipo de un parámetro de función: no puede ser el tipo de una variable, de un retorno ni de un campo (`Type 'session<Cuenta>' can only be used for function parameters`).
  - Una variable de tipo `session<S>` no se puede asignar, y dos sesiones no se pueden comparar.

Además, un método de un servicio puede no tener tipo de retorno, como `func reiniciar();`, y se dice que devuelve `Void`.

### Importación de módulos e interfaces

El diseño proponía, como requisito mínimo, importar solo archivos de interfaz, y planteaba la posibilidad de una extensión distinta para ellos. Se implementó la versión general (ver [Importaciones en Ricol](imports.md)):

- **Se puede importar cualquier archivo:** Un módulo es un archivo `.ric` común, que puede tener cualquier sentencia, no solo declaraciones. Sus sentencias se ejecutan una sola vez, antes que las del archivo que lo importa.
- **Al principio del archivo:** los imports van antes de cualquier otra sentencia. Un import en otro lugar es un error.
- **Rutas:** Como en el diseño, se aceptan tanto las rutas relativas como las rutas absolutas.
- **Una sola carga por archivo:** Como en el diseño, el mismo archivo importado por caminos distintos no duplica las declaraciones. Se lo reconoce por su ruta absoluta.
- **Ciclos:** Un ciclo de imports es un error previo a la ejecución, como en el diseño.
- **Nombres:** como en el diseño, todo lo declarado en el scope global de un módulo es visible en el archivo que lo importa, y dos declaraciones con el mismo nombre en archivos distintos son un error. Además, la visibilidad se transmite: si `main.ric` importa `a.ric` y `a.ric` importa `b.ric`, `main.ric` también ve lo declarado en `b.ric`.
- **Errores con archivo:** como un programa puede estar en varios archivos, todos los errores, de cualquier fase, indican el archivo que lo ocasiona: `[cuenta.ric, line 3, column 8] Undefined type 'Intts'`.

### Servicios y servidores

Lo central se implementó como en el diseño:
- cada método del `service` tiene que estar implementado por una función del `server` con el mismo nombre y la misma firma;
- las demás funciones del servidor son privadas;
- un programa declara como mucho un servidor;
- al terminar sus sentencias, el programa escucha en la dirección del servidor hasta que se lo interrumpe (Ctrl+C);
- el servidor atiende las llamadas de a una, y las sesiones de distintos clientes se intercalan llamada por llamada.

Las reglas que se agregaron son (ver[Servicios y servidores en Ricol](services.md)):

- **Declaración de un servicio:**
  - Los servicios y los servidores solo se pueden declarar en el nivel superior del programa, no dentro de un bloque, de una función ni de otro servidor.
  - Un servicio tiene que tener al menos un método, y no puede repetir métodos ni parámetros dentro de un método.
  - Un servicio no es un tipo ni un valor: su nombre solo se usa en `server`, `connect`, `connection<S>` y `session<S>`.
- **Implementación de los métodos:**
  - Los nombres de los parámetros enviados no tienen que coincidir con los del servicio. Solo sus tipos, en el mismo orden.
  - Las funciones que implementan los métodos tienen que estar declaradas directamente en el cuerpo del servidor, no dentro de un bloque.
  - Los tipos se comparan por su declaración: un struct declarado dentro del servidor con el mismo nombre que uno de afuera es otro tipo.
- **El cuerpo del servidor:**
  - Es un scope, como un bloque: lo declarado adentro no se ve después del servidor.
  - Se ejecuta cuando el programa llega a la declaración, y después el programa sigue con la sentencia siguiente. Recién al terminar todo el programa empieza a escuchar.
- **Errores al ejecutar:**
  - Si un método tiene un error local, como una división por cero, el cliente recibe un `ServerError` con el mensaje del error. El servidor lo escribe en la salida de errores y sigue atendiendo llamadas del mismo u otros clientes.
  - Si el servidor no puede escuchar, por ejemplo porque el puerto está en uso, es un error local con la posición de la dirección: `[servidor.ric, line 3, column 18] Cannot listen on 0.0.0.0:3825: address already in use`.

Más detalles sobre cómo se ejecuta el servidor están en [El servidor](network.md#el-servidor).

### Conexiones y sesiones

`connect` se implementó al igual que en el diseño: describe la conexión con un servidor, sin abrirla. La sesión cambió de sintaxis (ver [Ejemplo introductorio](#ejemplo-introductorio)), pero su ejecución sigue los pasos del diseño: se abre la conexión con un handshake que verifica el hash de la interfaz, se ejecuta el cuerpo, cada llamada espera su respuesta, y al terminar se cierra la conexión. Si algo falla, se ejecuta el `rescue`, que es obligatorio.

Lo que se definió al implementarla (ver [Conexiones y sesiones en Ricol](sessions.md)):

- **La conexión:** puede ser cualquier expresión de tipo `connection<S>`, no solo una variable: una llamada, un `connect` o el campo de un struct.
- **El timeout:**
  - Es opcional y es cualquier expresión de tipo `Int`, en milisegundos. Si no está, se usan 5000.
  - Que no sea positivo, o que supere el máximo representable (9223372036854 ms), es un error local que aborta el programa sin pasar por `rescue`.
  - Como en el diseño, el límite se aplica por separado a la apertura y a cada llamada. Si se agota durante la apertura, la falla es `ConnectionFailed`, no `Timeout`, porque se sabe que no se ejecutó nada.
- **Formas de salir del cuerpo:** además de terminar normalmente o con `return`, como decía el diseño, el cuerpo puede salir con `break` o `continue`. En todos los casos se cierra la conexión.
- **Sesiones anidadas:** cada falla se maneja en el `rescue` de la sesión a la que pertenece. Si falla una llamada de la sesión de afuera dentro del cuerpo de la de adentro, se cierran las dos y se ejecuta solo el `rescue` de afuera.
- **Scopes:** el nombre de la sesión y el cuerpo forman un scope, y el nombre de la falla y el `rescue` forman otro. Por eso el `rescue` no ve la sesión ni lo declarado en el cuerpo.
- **`return`:** en una función con tipo de retorno, una sesión cuenta como un `return` en todos los caminos solo si el cuerpo y el `rescue` retornan los dos.
- **Una conexión por sesión:** cada sesión abre su propia conexión TCP y la cierra al terminar. No se reutilizan conexiones entre sesiones.

### Errores

Las tres clases de errores del diseño se mantienen: los errores de negocio son valores de un enum que devuelve el servidor, las fallas de la sesión van al `rescue` y los errores locales abortan el programa. `SessionError` tiene las mismas cinco variantes que en el diseño, y está declarado en el scope global, así que no se puede declarar otra cosa con ese nombre en el nivel superior.

Lo que el diseño no definía es qué situación produce cada variante. Se clasifican según lo que el cliente sabe sobre lo que pasó en el servidor, como proponía la motivación del diseño (ver [Clasificación de las fallas](network.md#clasificación-de-las-fallas)):

| Falla | Cuándo | ¿Se ejecutó el método? |
| --- | --- | --- |
| `ConnectionFailed` | No se pudo conectar, o se agotó el tiempo durante la apertura | No |
| `ProtocolError` | El servidor implementa otra interfaz, no es un programa Ricol o envió un mensaje inválido | No, si es durante la apertura |
| `Timeout` | No llegó la respuesta de una llamada a tiempo | No se sabe |
| `ConnectionLost` | Se cortó la conexión durante una llamada | No se sabe |
| `ServerError` | El método tuvo un error local en el servidor | Sí, hasta el error |

Dos casos que surgieron al implementarlo:

- **Argumentos que no se pueden enviar:** si un argumento no se puede serializar (por ejemplo, un valor con demasiados niveles de anidamiento), es un error local en la posición de la llamada, y no se envía nada al servidor. No es una falla de la sesión, porque el problema está en el cliente.
- **Ejecución después de un `Timeout`:** con la semántica _at most once_, un método que falló con `Timeout` puede haberse ejecutado o no. Además, puede ejecutarse **después** de la falla: si la llamada esperaba su turno detrás de una llamada lenta, el servidor la ejecuta cuando le llega el turno, aunque el cliente ya no espere la respuesta.

### `match`

`match` se implementó como en el diseño, con la expresión entre paréntesis (ver [Ejemplo introductorio](#ejemplo-introductorio)). Las reglas que se agregaron son (ver [`match`](enums.md#match)):

- **Exhaustividad:** un `match` tiene que cubrir todas las variantes del enum o tener una rama `_` (`Non-exhaustive match: missing variant 'Rectangulo' of enum 'Forma'`).
- **Ramas inalcanzables:** son un error. Es el caso de una rama después de `_`, de una variante repetida, o de un `_` cuando ya están cubiertas todas las variantes.
- **Patrones:** cada campo de la variante se liga a un nombre, en orden, y `_` en lugar de un nombre ignora ese campo. No hay patrones anidados ni con literales.
- **Solo enums:** solo se puede hacer `match` de un valor de un enum (`Cannot match on type Int`).
- **`return`:** si todas las ramas terminan con `return`, el `match` cuenta como un `return` en todos los caminos.

### Características de la implementación del intérprete

#### Tipado estático y fuerte

Se implementó como en el diseño. El ejemplo de `c.retirar("30")` es un error de tipos, que se informa antes de ejecutar el cliente y sin necesidad del servidor:

```plaintext
Argument 1 of method 'retirar' of service 'Cuenta' must be Int, got String
```

La única conversión implícita es de `Int` a `Float`, como decía el diseño. Se aplica en las operaciones aritméticas y también en las comparaciones: `1 == 1.0` es `True`.

#### Fases del intérprete

El diseño describía cuatro fases. Con la carga de módulos son cinco, encadenadas en `src/ricol/ricol.go`:

```go
tokens, err := scanner.NewScanner(fileContent, r.filePath).Scan()
...
statements, err := parser.NewParser(tokens).Parse()
...
modules, err := loader.NewLoader(r.filePath, statements, os.ReadFile).Load()
...
checker := typechecker.NewTypeChecker(modules)
distances, checkErrors := checker.Check()
if len(checkErrors) > 0 {
    return checkErrors
}
...
program := interpreter.NewInterpreter(slices.Concat(modules...), distances, r.output)
program.EnableNetwork(interfacesOf(checker.Services()))
program.SetInput(r.input)
if err := program.Interpret(); err != nil {
    return err
}
return program.Serve()
```

1. **Análisis léxico** (scanner), como en el diseño. Cada token guarda, además de su línea y su columna, el archivo.
2. **Análisis sintáctico** (parser), como en el diseño.
3. **Carga de módulos** (loader): la fase nueva. Recorre los imports, escanea y parsea cada módulo, y arma un único programa (ver [Carga de módulos](imports.md#carga-de-módulos)).
4. **Verificación de tipos** (typechecker): además de verificar los tipos, resuelve a qué declaración se refiere cada nombre. Devuelve, para cada uso de un nombre (una variable, una función, un struct o un enum), a cuántos scopes de distancia está su declaración, y el intérprete usa esas distancias para buscarla.
5. **Ejecución** (interpreter): ejecuta el programa y, si declaró un servidor, escucha con `Serve` al terminar.

Los flags `--scan`, `--parse` y `--typecheck` detienen el proceso después de la fase correspondiente, para inspeccionar su resultado (ver [Ejecución](../README.md#ejecución)).

La separación entre los nodos del AST, que son solo datos, y los recorridos del chequeador y del intérprete, se mantiene como en el diseño: `checkExpression` devuelve un `types.Type` y `evaluate`, un `types.Value`.

#### Formato en la red y hash de la interfaz

El formato se implementó como en el diseño: los valores viajan sin nombres ni etiquetas de tipo, un `Int` ocupa 8 bytes, un struct es la secuencia de sus campos y un enum es el índice de su variante seguido de sus campos (ver [Formato en la red](wire.md)). Las diferencias están en el codec:

- **Una interfaz, no un par de funciones:** el codec es una interfaz de Go, y recibe la profundidad del valor:

```go
type codec interface {
    encode(writer io.Writer, value types.Value, depth int) error
    decode(reader io.Reader, depth int) (types.Value, error)
}
```

- **Profundidad máxima:** un valor puede tener hasta 100000 niveles de anidamiento. Sin el límite, un mensaje malicioso con un enum recursivo muy profundo agotaría el stack del servidor.
- **Datos inválidos:** los bytes mal formados nunca hacen fallar al intérprete: se informan con el error `ErrInvalidData`.
- **Tipos recursivos:** los codecs de un enum recursivo se arman una sola vez, y se reutilizan en sus propios campos.

El hash también se implementó como en el diseño: es el SHA-256 de una forma canónica del servicio y de los tipos que usa. La forma canónica empieza con una línea de versión (`ricol-interface-v1`), para que los hashes cambien si alguna vez cambia el formato de los valores (ver [Forma canónica y hash](wire.md#forma-canónica-y-hash)). Al abrir una sesión, el cliente envía el hash y el servidor lo compara con el suyo, como en el diseño (ver [Handshake](network.md#handshake)).

#### Fases futuras del intérprete

Las tres partes del plan del diseño se implementaron, con una diferencia en
dónde se generan los codecs y el hash:

- **Carga de módulos:** se implementó como estaba planeada, como una fase propia entre el parser y el typechecker. Cada posición registra también el archivo.
- **Codecs y hash:** el diseño planteaba que los generara el typechecker, que quizás se renombraría a análisis semántico. En cambio, están en un paquete aparte, `wire`, y el typechecker no se renombró. El typechecker solo expone los servicios que declaró (`Services()`), y después de una verificación sin errores se arma con `wire.NewInterface` la interfaz de cada servicio, con sus codecs y su hash. Se arma después, y no durante la verificación, porque los codecs necesitan que todos los tipos estén resueltos y sean válidos.
- **Runtime de red:** está en un paquete aparte, `network`, que no depende del intérprete (ver [Runtime de red en Ricol](network.md)). El intérprete ejecuta las sesiones a través de una interfaz de transporte, y el transporte de red la implementa con ese paquete.

### Agregados que no estaban en el diseño

- **Entrada por teclado:** `input()` es una expresión que lee una línea de la entrada estándar y la devuelve como `String`, sin el salto de línea final. Si no queda nada por leer, es un error local: `End of input` (ver [Entrada por teclado](../README.md#entrada-por-teclado)).
- **Protocolo de red:** el diseño solo mencionaba el handshake con el hash. El protocolo real tiene un saludo de tamaño fijo (38 bytes), frames con el largo de cada mensaje y un byte de estado en cada respuesta. Un servidor que rechaza la interfaz responde con un mensaje que nombra el servicio que implementa (ver [Protocolo](network.md#protocolo)).
- **Log del servidor:** cada método que falla se escribe en la salida de errores del servidor, por ejemplo: `Method 'retirar' of server 'Cuenta' failed: [servidor.ric, line 18, column 25] Cannot divide by zero: 10 // 0`.

## Diferencias con Lox

El intérprete de Ricol está escrito desde cero en Go, pero sigue la misma arquitectura que el intérprete de Lox: un scanner, un parser recursivo descendente, un paso estático que resuelve los nombres y un intérprete que recorre el AST con una cadena de environments.

Las diferencias están en el lenguaje. Ricol agrega tipado estático, un sistema de importaciones, structs y enums en lugar de clases, el sistema de RPC y la entrada por teclado.

### Tipado estático

Lox es de tipado dinámico: el tipo de cada valor se verifica al ejecutar cada operación, así que un error de tipos aparece recién cuando el programa llega a esa línea. Ricol es de tipado estático y fuerte (ver [Tipado estático y fuerte](#tipado-estático-y-fuerte)):

- **Tipos explícitos:** las variables, los parámetros, los retornos y los campos llevan su tipo. Todos los errores de tipos se informan juntos, antes de ejecutar ninguna sentencia.
- **Dos tipos de número:** Lox tiene un único tipo de número, de punto flotante. Ricol distingue `Int` (entero de 64 bits con signo) de `Float`, y la única conversión implícita es de `Int` a `Float`.
- **Sin truthiness:** en Lox, `nil` y `false` son falsos y cualquier otro valor es verdadero. En Ricol, las condiciones de `if` y `while`, y los operandos de `not`, `and` y `or`, tienen que ser `Bool` (`Non boolean expression in if condition: Int`).
- **Sin `nil`:** En Ricol no existe `nil`, así que el valor inicial es obligatorio.
- **Funciones que no son valores:** En Ricol una función solo se puede llamar (`Cannot use function 'f' as a value`). Una función declarada dentro de otra sí ve las variables de los scopes que la rodean, pero no puede salir de su scope.

Ricol resuelve todos los nombres, también los globales, y cada nombre es visible desde su declaración. Por eso Ricol necesita los grupos de funciones: las declaraciones `func` seguidas se ven entre sí, para permitir la recursión mutua (ver [Desde dónde es visible un nombre](scopes_and_names.md#desde-dónde-es-visible-un-nombre)).

### Sistema de importaciones

Ricol agrega `import` y una fase nueva del intérprete, la carga de módulos, que arma un único programa con todos los archivos antes de verificar los tipos (ver [Importación de módulos e interfaces](#importación-de-módulos-e-interfaces)). Es lo que permite que un cliente y un servidor compartan la interfaz de un servicio. Como un programa puede estar en varios archivos, los errores indican en cuál ocurrieron.

### Enums y structs

Para armar datos compuestos, Lox tiene clases y programacion orientada a objetos. Ricol no tiene clases. En su lugar tiene:

- **Structs:** Registros con campos fijos, declarados con su tipo, e inmutables. Dos structs se comparan campo por campo.
- **Enums:** Tipos cuyos valores son una entre varias variantes, cada una con sus propios campos. Sus valores se analizan con `match`, que tiene que cubrir todas las variantes.
- **Funciones en lugar de métodos:** el comportamiento está en funciones comunes, que reciben los valores como parámetros.

La inmutabilidad y la falta de métodos permite copiar un struct o un enum de un programa a otro, al ser los tipos que se envian por la red. Además, los enums recursivos ocupan el lugar de las estructuras enlazadas.

### Sistema RPC

Lox no tiene ninguna forma de comunicarse con otro programa, así que todo el sistema de RPC de Ricol es nuevo.

## Ejemplo integrador: ta-te-ti

En `examples/tictactoe/` hay un ta-te-ti en red: el jugador, desde el cliente,
juega contra la máquina, que es el servidor. Es el ejemplo más completo del
repositorio, porque usa juntas casi todas las partes de Ricol descritas en este
documento.

El servicio tiene tres métodos:

```ricol
service TicTacToe {
    func newGame(machineStarts: Bool) -> Started;
    func play(game: Int, cell: Int) -> Answer;
    func abandon(game: Int);
}
```

Está repartido en seis archivos:

- `tictactoe.ric`: la interfaz, con el servicio y los tipos que usa.
- `game.ric`: la lógica del juego, que usa solo el servidor: el tablero, el
  ganador y la jugada de la máquina.
- `logging.ric`: el registro de cada partida en la salida del servidor.
- `server.ric`: el servidor, en `0.0.0.0:3840`.
- `view.ric`: lo que usa solo el cliente: cómo se muestra el tablero y cómo se
  interpreta la jugada que escribe el jugador.
- `client.ric`: el cliente, con las preguntas al jugador y la sesión.

Lo que muestra de cada parte del lenguaje:

- **Imports:** el servidor y el cliente importan la misma interfaz, cada uno a través de sus propios módulos (`game.ric` y `view.ric`). El cliente no carga la lógica del juego.
- **Enums recursivos en lugar de listas:** Ricol no tiene arrays, así que el tablero es un enum recursivo con una casilla por variante `Square`:

  ```ricol
  enum Board {
      End;
      Square(cell: Cell, rest: Board);
  }
  ```

  Las funciones que leen o cambian una casilla recorren el tablero con `match` y recursión. El estado del servidor también es un enum recursivo: `Games` es la lista de las partidas en curso, cada una con su número y su tablero.
- **Resultados de negocio contra fallas de comunicación:** `play` devuelve un `Answer`, que tiene una variante para la jugada hecha (`Played`, con el tablero nuevo y el estado de la partida) y otras para los errores del juego (`CellTaken`, `InvalidCell` y `UnknownGame`). El cliente los analiza con `match`, como cualquier valor. Las fallas de comunicación, en cambio, solo llegan al `rescue` de la sesión, que tiene un mensaje para cada variante de `SessionError`.
- **Una sesión para todo el juego:** el cliente abre una única sesión, con `timeout 2000`, y juega todas las partidas dentro de ella. La función `playGame` recibe la sesión como parámetro (`machine: session<TicTacToe>`), que es el único lugar donde puede aparecer ese tipo.
- **`input()`:** el cliente lee las respuestas y las jugadas del jugador. Como Ricol no tiene una conversión de `String` a `Int`, la jugada se interpreta comparando el texto con `"1"` a `"9"` y con `"q"`, y se convierte en un enum `Choice` (`Move(cell)`, `Quit` o `Invalid`).
- **El servidor:** la máquina maneja la logica interna mediante if's. Varios clientes pueden jugar a la vez, cada uno con su partida: el estado es uno solo, y las llamadas de los distintos clientes se atienden de a una. El servidor registra cada jugada con `print`:

  ```plaintext
  [Juego 1] Partida nueva, empieza el jugador
  [Juego 1] El jugador jugó en la celda 1
  [Juego 1] La máquina jugó en la celda 5
  [Juego 1] El jugador jugó en la celda 2
  [Juego 1] La máquina jugó en la celda 3
  [Juego 1] El jugador eligió una celda ocupada
  [Juego 1] El jugador jugó en la celda 4
  [Juego 1] La máquina jugó en la celda 7
  [Juego 1] La máquina ganó
  ```

Para probarlo, desde `src/`, se levanta el servidor en una terminal y se ejecuta
el cliente en otra. Se pueden abrir varios clientes a la vez:

```sh
go run . ../examples/tictactoe/server.ric
go run . ../examples/tictactoe/client.ric
```
