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

(completar)

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

Servidor (`servidor.ric`):

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

Cliente (`cliente.ric`):

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
del mismo lenguaje (en el archivo de interfaz), y cómo el código tanto del
servidor como dentro del bloque `session` del cliente son completamente
agnósticos a la distribución del sistema (las llamadas a las funciones de la
`Cuenta` se realizan de forma idéntica a llamadas locales).

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
`server`, empieza a escuchar en su dirección hasya que se lo interrumpe (con
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

(completar)
