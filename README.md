# Ricol

Ricol es un lenguaje diseñado para facilitar el uso de Remote Procedure Calls
(RPC).

El diseño e implementación de Ricol forma parte del Trabajo Práctico de la
materia Lenguajes y Compiladores I, cátedra de
[Federico del Mazo](https://github.com/fdelmazo/).

Integrantes del grupo:

- [Agustín Altamirano](https://github.com/AgustinAltamirano), padrón 110237
- [Santiago Fassio](https://github.com/SantiagoFassio), padrón 109463

## Índice

- [Ejecución](#ejecución)
- [Sintaxis de Ricol](#sintaxis-de-ricol)
  - [Hola mundo](#hola-mundo)
  - [Sentencias](#sentencias)
  - [Tipos de datos](#tipos-de-datos)
  - [Comentarios](#comentarios)
  - [Operaciones aritméticas](#operaciones-aritméticas)
  - [Concatenación de strings](#concatenación-de-strings)
- [Motivación de Ricol](#motivación-de-ricol)
- [Lenguajes y tecnologías precedentes](#lenguajes-y-tecnologías-precedentes)
- [Sintaxis propuesta de manejo de Remote Procedure Calls](#sintaxis-propuesta-de-manejo-de-remote-procedure-calls)
  - [Ejemplo introductorio](#ejemplo-introductorio)
  - [Tipos nuevos](#tipos-nuevos)
  - [Importación de módulos e interfaces](#importación-de-módulos-e-interfaces)
  - [Servicios y servidores](#servicios-y-servidores)
  - [Conexiones y sesiones](#conexiones-y-sesiones)
  - [Errores](#errores)
  - [`match`](#match)
  - [Comparación con otros lenguajes](#comparación-con-otros-lenguajes)
- [Características de la implementación del intérprete](#características-de-la-implementación-del-intérprete)
  - [Tipado estático y fuerte](#tipado-estático-y-fuerte)
  - [Fases del intérprete](#fases-del-intérprete)
  - [Formato en la red y hash de la interfaz](#formato-en-la-red-y-hash-de-la-interfaz)
  - [Fases futuras del intérprete](#fases-futuras-del-intérprete)

## Ejecución

El intérprete de Ricol está escrito en Go y requiere Go 1.27.1 o superior. Los
programas de Ricol se escriben en archivos con extensión `.ric`.

Desde el directorio `src/`, un programa se ejecuta con:

```sh
go run . ../examples/hello_world.ric
```

También se puede compilar el intérprete y ejecutarlo directamente:

```sh
go build -o ricol .
./ricol ../examples/hello_world.ric
```

El intérprete procesa el programa en cuatro fases: análisis léxico, análisis
sintáctico, verificación de tipos y ejecución. Los siguientes flags permiten
detener el proceso luego de una de las fases, para inspeccionar su resultado:

| Flag | Resultado |
| --- | --- |
| `--scan` | Imprime los tokens reconocidos |
| `--parse` | Imprime las sentencias, con las expresiones agrupadas según su precedencia |
| `--typecheck` | Verifica los tipos, sin ejecutar el programa |

Los tests del intérprete se ejecutan desde `src/` con `go test ./...`.

## Sintaxis de Ricol

La gramática formal de Ricol se encuentra en
[Backus-Naur Form de Ricol](docs/bnf.md).

### Hola mundo

```ricol
PRINT "Hola mundo";
```

La palabra reservada `PRINT` permite imprimir valores por pantalla. Puede
imprimir valores de cualquier tipo, no solo strings. No agrega salto de línea al
final.

### Sentencias

Un programa en Ricol está formado por una secuencia de sentencias. Las
sentencias deben finalizar en `;` y pueden ocupar varias líneas.

### Tipos de datos

Ricol es un lenguaje de tipado **estático** y **fuerte**. Posee los siguientes
tipos de datos primitivos:

- **Int**: números enteros de 64 bits, con signo. Ejemplo: `3`.
- **Float**: números de punto flotante de 64 bits según lo establecido por el
estándar IEEE 754, con la excepción de que dividir por cero es un error.
Ejemplo: `41.3`.
- **String**: cadenas de carácteres. Ejemplo: `"Ricol"`.

Que el tipado sea **estático** significa que los tipos de todas las expresiones
se verifican antes de ejecutar el programa. Si hay algún error de tipos, se
informan todos juntos y no se ejecuta ninguna sentencia. Por ejemplo, el
siguiente programa no imprime nada:

```ricol
PRINT "Hola";
"a" - "b";      @ Error de tipos: los strings no se pueden restar
```

Que sea **fuerte** significa que no hay conversiones implícitas entre tipos,
excepto de `Int` a `Float` en las operaciones aritméticas.

Más información en la documentación de los tipos:

- [Números](docs/numbers.md)
- [String](docs/strings.md)

### Comentarios

```ricol
@ Línea con un comentario

PRINT "Ricol"; @ Comentario inline
```

Los comentarios pueden colocarse al inicio de una nueva línea del código o de
forma _inline_, al finalizar una sentencia.

### Operaciones aritméticas

```ricol
1 + 2.3;
2 * 3.1 - 5;
2 * (3.1 + 5);
10 / 3;         @ División flotante (=3.3333)
10 // 3;        @ División entera con redondeo hacia abajo (=3)
3 % 2;          @ Resto (=1)
-(-95.4);       @ =95.4
4 ** 2;         @ Potencia =16
```

Ricol posee operadores para realizar las operaciones aritméticas más comunes:
suma, resta, multiplicación, división y potenciación, además del uso del
operador `-` para obtener el opuesto de un número. También define el operador
`//` para la división entera con redondeo hacia abajo y el `%` para la operación
resto (_modulo_ en inglés). Al igual que en otros lenguajes de programación (por
ejemplo Python), el resto tiene el signo del divisor: `-7 // 2` es `-4` y
`-7 % 2` es `1`.

Las operaciones aritméticas pueden realizarse con operandos de tipo `Int` y
`Float` (por ejemplo, en Ricol se puede evaluar `1 + 2.3`). El tipo del
resultado depende de la operación y de sus operandos. Como regla general, si
todos los operandos son `Int`, la operación resultará en un `Int` (excepto la
división, que siempre devuelve `Float`). Si algún operando es un `Float`, la
operación resulta en un `Float`.

Estas operaciones siguen las reglas de precedencia de la aritmética. Ricol
permite el uso de `()` para agrupar operaciones.

Para saber más acerca de estas operaciones, su precedencia y sus tipos de
retorno se puede consultar el documento
[Operaciones aritméticas en Ricol](docs/arithmetic.md).

### Concatenación de strings

```ricol
PRINT "Hola, " + "mundo";   @ Imprime Hola, mundo
```

El operador `+` concatena dos strings. Es el único operador que admiten los
strings, y ambos operandos deben ser strings.

Más información en [Strings en Ricol](docs/strings.md).

## Motivación de Ricol

El término _Remote Procedure Call_ (RPC) se atribuye a Bruce Jay Nelson, quien
lo formalizó en su tesis doctoral (_Remote Procedure Call_, Carnegie Mellon,
1981). En ella lo define como:

> la transferencia síncrona de control, a nivel de lenguaje, entre programas en
> espacios de direcciones disjuntos, cuyo medio principal de comunicación es un
> canal angosto.

Tres años después, Andrew Birrell y Nelson publicaron
[_Implementing Remote Procedure Calls_](https://web.eecs.umich.edu/~mosharaf/Readings/RPC.pdf)
(ACM TOCS, 1984), donde describen conceptos que siguen vigentes hoy en día:
los stubs del lado cliente y servidor, el empaquetado (_packing_) y
desempaquetado (_unpacking_) de argumentos y resultados, el _binding_ entre
cliente y servidor, y la semántica conocida como _at most once_ (si la llamada
vuelve, el procedimiento se ejecutó exactamente una vez; si falla, puede
haberse ejecutado una vez o ninguna).

De esta forma, las _Remote Procedure Calls_ se plantean como una capa de
abstracción por encima del transporte de red, en la cual se le otorga una
semántica especial al envío de un mensaje cliente --> servidor (y su
correspondiente respuesta) para que se asemeje a una llamada a un procedimiento
o función local.

Sin embargo, en la práctica la programación con RPCs tiene ciertas asperezas,
las cuales dependen del lenguaje, el protocolo y las bibliotecas utilizadas. En
muchos casos se dan las siguientes situaciones:

- La interfaz que expone el servidor y que consumen los clientes se debe
definir en un lenguaje aparte (un _Interface Definition Language_, como
Protocol Buffers en gRPC), distinto del lenguaje en el que se programan cliente
y servidor. Esto requiere un paso extra de generación de código y herramientas
adicionales, y el código generado puede quedar desactualizado respecto de la
interfaz si no se lo regenera.
- Los tipos utilizados en las llamadas no son los mismos que los del lenguaje
de programación del cliente y del servidor, sino los que genera la
herramienta. Se requiere un mapeo o conversión entre ellos en el código.
- El manejo de las fallas de red o de conexión se esparce por cada llamada
remota (un `try`/`catch` o un `if err != nil` después de cada una), lo cual
termina ensuciando el código de lógica de negocio. Además, los errores de
negocio y los de comunicación suelen llegar por el mismo canal (en gRPC, por
ejemplo, ambos tipos de errores se informan como un código de estado), y nada
obliga al programador a manejar las fallas.

En base a todo esto, Ricol se plantea con el objetivo de facilitar el uso de
RPCs y volverlas idiomáticas al lenguaje en sí.

Algo importante a tener en cuenta es el paper de Waldo, Wyant, Wollrath y
Kendall, [_A Note on Distributed Computing_](https://waldo.scholars.harvard.edu/sites/g/files/omnuum6261/files/waldo/files/waldo-94.pdf)
(Sun Microsystems, 1994), en el cual se argumenta que los sistemas que ocultan
la diferencia entre objetos locales y remotos fracasan, porque la latencia, el
modelo de acceso a memoria, la concurrencia y las fallas parciales no se pueden
esconder. Teniendo eso en mente, Ricol busca el equilibrio: simplificar el uso
de llamados remotos para el usuario, pero sin ocultar la distribución. En vez de
marcar cada llamada remota, concentra toda la diferencia entre lo local y lo
remoto en un único lugar, el bloque `session` / `rescue`.

Ricol no resuelve todos los problemas que plantea el paper: se enfoca en las
fallas parciales. La latencia de cada llamada sigue estando presente dentro de
una sesión, y los problemas de concurrencia se evitan haciendo que el servidor
atienda una llamada por vez.

Concretamente, Ricol propone:

- **Contratos escritos en el mismo lenguaje**: la interfaz de un servicio se
define en un archivo Ricol, con los mismos tipos que usan cliente y servidor,
sin generación de código. El chequeo de tipos verifica, antes de ejecutar, que
el servidor implemente el contrato y que el cliente lo use correctamente.
- **Separación entre errores de negocio y fallas de comunicación**: los errores
de negocio son parte del tipo de retorno de cada función del servicio. Las
fallas de comunicación se manejan en un único bloque `rescue`, clasificadas
según lo que el cliente sabe sobre lo que pasó en el servidor (por ejemplo, si
no se pudo conectar no se ejecutó nada, pero si se cortó la conexión durante
una llamada no se sabe si se ejecutó).
- **Verificación de compatibilidad:** al abrir una sesión, cliente y servidor
comparan un hash de la interfaz, por lo que una versión incompatible se detecta
antes de intercambiar datos.

Estas garantías son las que justifican diseñar un lenguaje en lugar de una
biblioteca: una biblioteca no puede verificar antes de la ejecución que una
sesión no escape de su bloque ni que toda llamada remota tenga su manejo de
fallas.

## Lenguajes y tecnologías precedentes

(completar)

## Sintaxis propuesta de manejo de Remote Procedure Calls

### Ejemplo introductorio

Para entender bien la sintaxis propuesta de Ricol para el manejo de RPCs, se
plantea el siguiente ejemplo de comunicación entre cliente y servidor.

Se cuenta con tres archivos: la interfaz, el servidor y el cliente.

**Interfaz**, compartida por cliente y servidor (`cuenta.ric`):

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

server Cuenta at "tcp://0.0.0.0:3825" {
    let actual: Int = 0;

    func saldo() -> Int {
        return actual;
    }

    func depositar(monto: Int) -> Movimiento {
        actual = actual + monto;
        return Movimiento { monto: monto, saldoFinal: actual };
    }

    func retirar(monto: Int) -> ResultadoRetiro {
        if monto > actual {
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

let conn: connection<Cuenta> = connect Cuenta at "tcp://localhost:3825";

func retirarYMostrar(c: session<Cuenta>, monto: Int) -> Bool {
    match c.retirar(monto) {
        Ok(mov) => {
            print "retiro ok, saldo: ", mov.saldoFinal, "\n";
            return true;
        }
        SaldoInsuficiente(disp) => {
            print "saldo insuficiente, disponible: ", disp, "\n";
            return false;
        }
    }
}

session c = conn timeout 2000 {
    c.depositar(100);
    retirarYMostrar(c, 30);
    retirarYMostrar(c, 500);
    print "saldo final: ", c.saldo(), "\n";
} rescue e {
    match e {
        ConnectionFailed(msg) => print "no se pudo conectar: ", msg, "\n";
        ConnectionLost(msg)   => print "se cortó la conexión, no se sabe si la última operación se hizo: ", msg, "\n";
        Timeout               => print "el servidor no respondió a tiempo\n";
        ProtocolError(msg)    => print "cliente y servidor no son compatibles: ", msg, "\n";
        ServerError(msg)      => print "error interno del servidor: ", msg, "\n";
    }
}
```

Obsérvese cómo la definición del contrato del servicio remoto se realiza dentro
del mismo lenguaje (en el archivo de interfaz), y cómo dentro del bloque
`session` del cliente las llamadas a las funciones de la `Cuenta` se escriben
igual que llamadas locales. Toda la diferencia con lo local (abrir y cerrar la
conexión, el tiempo de espera y las fallas de comunicación) se concentra en el
bloque `session` / `rescue`.

### Tipos nuevos

Los tipos ya existentes en el lenguaje (`Int`, `Float`, `String` y `Bool`) son
**transportables**, es decir, pueden ser utilizados para los argumentos en las
llamadas a funciones remotas o para sus valores de retorno.

Además de esos, se agregarían los siguientes tipos:

- **`struct D`**: registros que poseen campos con nombre. Es inmutable, por lo
que pasarlos por copia o por referencia da lo mismo. Son transportables si todos
sus campos lo son.
- **`enum E`**: tipo enum que permite una entre varias variantes especificadas.
Las variantes pueden tener campos opcionales. Al igual que los `struct`, son
inmutables. Son transportables si todos los campos de las variantes lo son.
- **`connection<S>`**: tipo de dato que describe la conexión a un servicio `S`
(indica qué servicio es y cuál es su dirección). No es transportable.
- **`session<S>`**: representa una sesión abierta con un servicio `S`. No es
transportable, y nunca puede escapar de su correspondiente bloque `session`.

### Importación de módulos e interfaces

Para tener definida la interfaz de un servicio en un archivo aparte, Ricol
necesita incorporar una sentencia especial para esto. En los ejemplos:

```ricol
import "cuenta.ric";
```

Se utilizarían rutas relativas al archivo a importar. A priori y como requisito
mínimo, se puede implementar la importación únicamente de archivos de interfaz
_(una variante podría ser definir una extensión distinta para estos archivos)_,
pero se puede extender para importar código Ricol de cualquier tipo.

Importar el mismo archivo dos veces por caminos distintos no duplicaría las
declaraciones, y un ciclo de imports sería considerado como un error previo a la
ejecución.

Los nombres de todo lo importado quedan visibles en el archivo que importa, y
dos declaraciones con el mismo nombre sería considerado como un error
(nuevamente, previo a la ejecución).

### Servicios y servidores

Un `service` declara un contrato, es decir, la lista de funciones con sus tipos
de parámetros y de retorno (los cuales deben ser transportables).

```ricol
service Cuenta {
    fun saldo() -> Int;
    fun depositar(monto: Int) -> Movimiento;
    fun retirar(monto: Int) -> ResultadoRetiro;
}
```

Un `server` implementa un servicio y lo expone en una dirección:

```ricol
server Cuenta at "tcp://0.0.0.0:3825" {
    let actual: Int = 0;                             @ estado global del servidor

    fun depositar(monto: Int) -> Movimiento { ... }  @ implementa un método del servicio

    fun validar(monto: Int) -> Bool {                @ auxiliar privada
        return monto > 0;
    }
}
```

Cada método del `service` tiene que estar implementado por una función con el
mismo nombre y la misma firma. Las funciones que no están en la interfaz son
privadas.

El programa servidor ejecuta sus sentencias en orden. Al terminar, si declaró un
`server`, empieza a escuchar en su dirección hasta que se lo interrumpe (con
Ctrl + C por ejemplo). Un programa declara como mucho un `server`.

En Ricol el programa servidor procesa las llamadas de a una y sin importar de
qué cliente vengan, lo cual evita las _race conditions_. Las sesiones de
distintos clientes se intercalan llamada por llamada.

### Conexiones y sesiones

La palabra reservada `connect` crea un descriptor de una conexión con un
`server`. No abre ninguna conexión en sí.

```ricol
let conn: connection<Cuenta> = connect Cuenta at "tcp://localhost:3825";
```

`session` abre una conexión, ejecuta el código dentro del bloque y la cierra:

```ricol
session c = conn timeout 2000 {
    ...                     @ c: session<Cuenta>
} rescue e {
    ...                     @ e: SessionError
}
```

La ejecución del programa sería algo como lo siguiente:

1. Se abre la conexión haciendo un `handshake` en el que se verifica el hash de
la interfaz. Si la conexión falla, se pasa al paso 5.
2. Se ejecuta el cuerpo de `session` con un nombre asociado al valor de tipo
`session<S>` (en el ejemplo `c`).
3. Cada llamada `c.funcion(argumentos)` envía el mensaje y espera la respuesta
antes de seguir. Tiene un límite de tiempo definido al crear la sesión
(en milisegundos), usando un valor por defecto en caso de no definirlo. Este
límite se aplica a cada llamada por separado y también a la apertura de la
conexión.
4. Si el cuerpo de `session` termina normalmente, o sale con `return`, se
cierra la conexión y la ejecución sigue según corresponda.
5. Si una llamada o una apertura de conexión falla, la ejecución del programa
abandona el cuerpo de `session`, se cierra la conexión y se ejecuta el bloque
`rescue`, con la variable ligada al `SessionError`.

Aclaración: un error local dentro del cuerpo (por ejemplo una división por cero
en el código del cliente) no va a `rescue`, sino que aborta el programa como en
cualquier otro código local.

El bloque `rescue` es obligatorio.

### Errores

Ricol distinguiría tres clases de errores:

- **Error de lógica de negocio**: en el ejemplo, el caso de saldo insuficiente.
El servidor los manifiesta como un valor de retorno de sus funciones, con un
tipo `enum`.
- **Falla en la sesión**: producida por cortes de red, interfaz que no coincide
entre cliente y servidor, falla de servidor, etc. Estos casos provocan que se
aborte la ejecución del cuerpo de `session` y se maneja con el bloque `rescue` y
su correspondiente `SessionError`.
- **Error local de ejecución**: abortan el programa.

`SessionError` es un `enum` predefinido:

```ricol
enum SessionError {
    ConnectionFailed(message: String);   @ no se pudo conectar
    ConnectionLost(message: String);     @ se cortó durante una llamada
    Timeout;                             @ no llegó respuesta a tiempo
    ProtocolError(message: String);      @ interfaz incompatible o mensaje inválido
    ServerError(message: String);        @ error interno del servidor al ejecutar la función
}
```

### `match`

Para manejar los casos de un tipo de dato `enum`, se usa la sentencia `match`:

```ricol
match r {
    Ok(mov)                 => print mov.saldoFinal;
    SaldoInsuficiente(disp) => { print "faltan fondos\n"; }
    _                       => print "otro caso\n";
}
```

### Comparación con otros lenguajes

(completar)

## Características de la implementación del intérprete

El intérprete de Ricol está escrito en Go y es un intérprete de tipo
_tree-walk interpreter_: el código fuente se transforma en un árbol de sintaxis
abstracta (AST) que luego se recorre para ejecutarlo. A continuación se detallan
algunas decisiones de diseño que se tomaron sobre su implementación,
relacionadas con el objetivo de Ricol.

### Tipado estático y fuerte

En una llamada remota intervienen dos programas que se ejecutan por separado,
posiblemente en máquinas distintas, y que solo comparten la interfaz del
servicio como contrato. Para que el contrato sirva, alguien tiene que verificar
que ambas partes lo cumplan.

Con tipado **dinámico**, esa verificación solo puede ocurrir durante la
ejecución. Por ejemplo, siguiendo los ejemplos vistos previamente, si un cliente
llama `c.retirar("30")`, con un `String` en lugar de un `Int`, el error recién
se detecta en tiempo de ejecución cuando el servidor intenta hacer alguna
operación con él (por ejemplo `monto > actual`). El cliente recibiría un error
de servidor, cuando en realidad el problema lo originó él y lo podría haber
evitado. Si en vez de fallar el servidor convirtiera el valor implícitamente
(como harían los lenguajes de tipado débil), cliente y servidor le estarían
dando significados distintos al mismo dato.

Con tipado **estático**, el mismo error se reporta antes de ejecutar el
cliente, sin necesidad de que el servidor esté levantado. El chequeador de
tipos puede verificar, a partir de un único archivo de interfaz escrito en
Ricol:

- que el servidor implemente cada función del servicio con la misma firma,
- que el cliente llame a esas funciones con argumentos del tipo correcto y use
  el resultado según su tipo de retorno,
- que todos los tipos que viajan por la red sean transportables.

Que el tipado sea además **fuerte** garantiza que un valor signifique lo mismo
de ambos lados: al no haber conversiones implícitas (salvo `Int` a `Float` en
las operaciones aritméticas), lo que el cliente envía como `Int` el servidor lo
recibe y lo usa como `Int`.

### Fases del intérprete

El intérprete procesa un programa en cuatro fases, encadenadas. Cada una recibe
el resultado de la anterior, y si falla, el programa no llega a la siguiente:

```go
tokens, err := scanner.NewScanner(fileContent).Scan()
...
statements, err := parser.NewParser(tokens).Parse()
...
checkErrors := typechecker.NewTypeChecker(statements).Check()
if len(checkErrors) > 0 {
    return checkErrors
}
...
err = interpreter.NewInterpreter(statements, r.output).Interpret()
```

1. **Análisis léxico** (scanning): transforma el texto del programa en
una secuencia de tokens. Cada token guarda su línea y su columna, que se usan
para ubicar los errores de todas las fases.
2. **Análisis sintáctico** (parsing): un parser recursivo _top-down_ construye
el AST a partir de los tokens, siguiendo la [gramática de Ricol](docs/bnf.md) y
sus reglas de precedencia.
3. **Verificación de tipos** (typechecker): recorre el AST y calcula el tipo de
cada expresión, sin ejecutar nada. Es el encargado de verificar que el tipado
estático se cumpla.
4. **Ejecución** (interpreter): recorre el AST y ejecuta cada sentencia.

Los nodos del AST no saben evaluarse ni chequearse a sí mismos: son solo datos.
El recorrido está en el chequeador y en el intérprete, que tienen la misma
estructura (un `switch` sobre el tipo de nodo) pero devuelven cosas distintas:

```go
// typechecker: obtiene el tipo de la expresión sin ejecutarla
func (t *TypeChecker) checkExpression(expression common.Expression) types.Type

// interpreter: obtiene el valor de la expresión ejecutándola
func (i *Interpreter) evaluate(expression common.Expression) (types.Value, error)
```

Esta separación es lo que permite que el tipado sea estático. Si solo existiera
`evaluate`, la única forma de conocer el tipo de una expresión sería
ejecutarla, que es justamente lo que sucede en el caso dinámico descrito antes.

### Formato en la red y hash de la interfaz

El tipado estático, además de detectar errores, aporta información que el
runtime de red va a usar.

Como cliente y servidor conocen, a partir de la interfaz, el tipo de cada
argumento y de cada valor de retorno, los mensajes no necesitan llevar nombres
de campos ni indicar de qué tipo es cada valor. Un `Int` ocupa siempre 8 bytes,
un `struct` es la secuencia de sus campos en orden de declaración y un `enum` es
el índice de su variante seguido de los campos de esa variante.

Para cada tipo transportable se arma, antes de ejecutar, un _codec_: un par de
funciones que codifican y decodifican valores de ese tipo. El codec de un
`struct` se construye componiendo los de sus campos:

```go
type codec struct {
    encode func(w io.Writer, value types.Value) error
    decode func(r io.Reader) (types.Value, error)
}
```

Es el equivalente al código de serialización que generan herramientas como gRPC.

**Hash de la interfaz.** El chequeador conoce la estructura completa de cada
servicio, así que puede calcular un hash (por ejemplo SHA-256) sobre una
representación canónica del servicio y de todos los tipos que usa.

Al abrir una sesión, el cliente envía el hash de su versión de la interfaz y el
servidor lo compara con el suyo. Si cliente y servidor se compilaron con
versiones distintas de la interfaz, la sesión falla con un `ProtocolError` en
lugar de interpretar mal los bytes recibidos. Esto es necesario precisamente
porque el formato no lleva etiquetas de tipo: sin el hash, un cambio en la
interfaz haría que un lado leyera datos con la estructura equivocada.

### Fases futuras del intérprete

Para soportar RPCs, el intérprete pasaría a tener las siguientes fases:

1. scanner
2. parser
3. module importer
4. typechecker (que suma la generación de codecs para los tipos y hashes para
las interfaces)
5. interpreter (con el agregado del runtime de red)

La fase nueva de carga de módulos (module importer) recorre los import del
programa, resolvería cada ruta relativa al archivo que la importa, y escanea y
parsea los archivos importados de forma recursiva. Detecta los ciclos de imports
y el mismo archivo importado por caminos distintos. El typechecker recibe un
único programa con las declaraciones de todos los módulos. Como los errores
pasan a poder estar en cualquier archivo, cada posición registra también el
archivo.

El typechecker pasaría a generar también los codecs y los hashes de cada
servicio. Debido a esto, es posible que esta fase se renombre a
análisis semántico o se separe en dos fases.
