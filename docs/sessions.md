# Conexiones y sesiones en Ricol

Un programa cliente se comunica con un servidor en dos pasos: `connect` describe la conexión con un servidor, y `session` la abre, ejecuta un bloque en el que los métodos del servicio se llaman como funciones locales y la cierra. Si la comunicación falla, se ejecuta el bloque `rescue`.

```ricol
import "cuenta.ric";

let conn: connection<Cuenta> = connect Cuenta at "localhost:3825";

session conn timeout 2000 as c {
    c.depositar(100);
    print c.saldo();
} rescue e {
    match (e) {
        ConnectionFailed(mensaje) => print "no se pudo conectar: " + mensaje;
        ConnectionLost(mensaje) => print "se cortó la conexión: " + mensaje;
        Timeout => print "el servidor no respondió a tiempo";
        ProtocolError(mensaje) => print "cliente y servidor no son compatibles: " + mensaje;
        ServerError(mensaje) => print "falló el servidor: " + mensaje;
    }
}
```

El servicio `Cuenta` es el de [Servicios y servidores en Ricol](services.md). Toda la diferencia con el código local (abrir y cerrar la conexión, el tiempo de espera y las fallas de comunicación) queda en el bloque `session` / `rescue`.

Los ejemplos de [Diseño de Ricol](design.md#conexiones-y-sesiones) usan una sintaxis preliminar (direcciones `tcp://`, `match` sin paréntesis, `print` con varios valores y el nombre de la sesión antes de la conexión, con `=`). En este documento se usa la sintaxis implementada.

## Estado actual: sin runtime de red

Por ahora el intérprete no tiene runtime de red: el servidor no escucha conexiones y el cliente no las abre. La verificación de tipos y la semántica de las sesiones están completas, pero al ejecutar un programa toda sesión falla al abrirse con `ConnectionFailed`. El ejemplo de arriba imprime:

```plaintext
no se pudo conectar: cannot connect to Cuenta at localhost:3825: network runtime not available
```

## Conexiones

### `connect`

`connect` es una expresión formada por la palabra reservada `connect`, el nombre del servicio, la palabra reservada `at` y la dirección del servidor. Su tipo es `connection<S>`, donde `S` es el servicio. No abre ninguna conexión: solo describe a qué servicio y a qué dirección conectarse.

El servicio tiene que estar declarado antes, y la dirección sigue las mismas reglas que la de un servidor (ver [Dirección](services.md#dirección)): es un literal de la forma `host:puerto`, que se verifica antes de ejecutar.

```ricol
let c: connection<Cuenta> = connect Fantasma at "tcp://localhost:3825";
@ Error: Undefined service 'Fantasma'
@ Error: Invalid address 'tcp://localhost:3825': expected host:port
```

### El tipo `connection<S>`

`connection<S>` es un tipo como cualquier otro: puede ser el tipo de una variable, de un parámetro, de un valor de retorno o de un campo de un struct o de una variante.

```ricol
struct Replica {
    nombre: String;
    destino: connection<Cuenta>;
}

func principal() -> connection<Cuenta> {
    return connect Cuenta at "localhost:3825";
}

print principal();  @ Imprime connection<Cuenta> at localhost:3825
print principal() == connect Cuenta at "[::1]:3826"; @ Imprime False
```

Dos conexiones se comparan con `==` y `!=`, que comparan el servicio y la dirección. No tienen orden ni admiten otros operadores.

La dirección se guarda normalizada: el puerto sin ceros a la izquierda y las direcciones IP en su forma canónica. Así, dos direcciones escritas de forma distinta son la misma si el puerto y la IP coinciden, y `print` muestra la forma normalizada. Los nombres de host no se resuelven ni se normalizan: `localhost` y `127.0.0.1` son direcciones distintas.

```ricol
print connect Cuenta at "localhost:3825" == connect Cuenta at "localhost:03825";  @ Imprime True
print connect Cuenta at "[::1]:3825" == connect Cuenta at "[0:0::1]:3825";        @ Imprime True
print connect Cuenta at "localhost:3825" == connect Cuenta at "127.0.0.1:3825";   @ Imprime False
print connect Cuenta at "[0:0::1]:03825";  @ Imprime connection<Cuenta> at [::1]:3825
```

Una conexión no tiene métodos: para llamar a un método hay que abrir una sesión.

```ricol
print conn.saldo();  @ Error: Cannot call 'saldo' on type connection<Cuenta>
```

El servicio de `connection<S>` tiene que estar declarado. Si `S` no existe o no es un servicio, el error es `Undefined service 'S'` o `'S' is not a service`.

## Sesiones

### Sintaxis

Una sesión es la palabra reservada `session`, la conexión, opcionalmente la palabra reservada `timeout` seguida del tiempo de espera, la palabra reservada `as`, el nombre de la sesión y el cuerpo entre `{}`. Después va la palabra reservada `rescue`, el nombre de la falla y el bloque que la maneja, entre `{}`. El bloque `rescue` es obligatorio:

```ricol
session conn as c {
    print c.saldo();
}     @ Error: Expected 'rescue' after session body
```

La conexión puede ser cualquier expresión de tipo `connection<S>`: una variable, una llamada, un `connect` o el campo de un struct.

```ricol
session 1 as c { } rescue e { }  @ Error: Cannot open a session on type Int
```

### Timeout

El timeout es una expresión opcional de tipo `Int`, en milisegundos. Es 5000 por defecto.

```ricol
session conn timeout 2.5 as c { } rescue e { }
@ Error: Session timeout must be Int, got Float
```

Que el timeout no sea positivo, o que supere los 9223372036854 milisegundos (unos 292 años, el máximo que se puede representar), es un error local, que se detecta al ejecutar la sesión y aborta el programa sin pasar por `rescue`:

```ricol
let espera: Int = 0;
session conn timeout espera as c { } rescue e { }
@ Error: Session timeout must be positive, got 0

session conn timeout 9223372036855 as c { } rescue e { }
@ Error: Session timeout must be at most 9223372036854, got 9223372036855
```

El límite se aplica por separado a la apertura de la conexión y a cada llamada.

### El cuerpo y las llamadas remotas

En el cuerpo, el nombre de la sesión (`c` en los ejemplos) es un valor de tipo `session<S>`. Una llamada `c.metodo(argumentos)` se verifica contra el servicio, como una llamada a una función, y su tipo es el tipo de retorno del método:

```ricol
session conn as c {
    let m: Movimiento = c.depositar(100);
    c.retirar();            @ Error: Method 'retirar' of service 'Cuenta' expects 1 argument, got 0
    c.depositar("100");     @ Error: Argument 1 of method 'depositar' of service 'Cuenta' must be Int, got String
    c.transferir(5);        @ Error: Service 'Cuenta' has no method 'transferir'
    print c.reiniciar();    @ Error: Cannot print a Void value
} rescue e { }
```

Cada llamada envía los argumentos al servidor y espera la respuesta antes de seguir. Los argumentos se evalúan de izquierda a derecha, igual que en una llamada local.

### `session<S>` no puede escapar de su bloque

Un valor de tipo `session<S>` solo existe mientras la sesión está abierta. Por eso `session<S>` solo puede ser el tipo de un parámetro de función, para escribir funciones auxiliares que usan la sesión:

```ricol
func depositarDos(c: session<Cuenta>, monto: Int) -> Int {
    c.depositar(monto);
    return c.depositar(monto).saldoFinal;
}

session conn as c {
    print depositarDos(c, 50);
} rescue e { }
```

En cualquier otro lugar es un error: en una variable, un tipo de retorno o un campo de un struct o de una variante.

```ricol
let s: session<Cuenta> = c;  @ Error: Type 'session<Cuenta>' can only be used for function parameters
```

Además, una variable de tipo `session<S>` (el nombre de la sesión o un parámetro) no se puede asignar, y dos sesiones no se pueden comparar:

```ricol
func f(a: session<Cuenta>, b: session<Cuenta>) {
    a = b;                  @ Error: Cannot assign to session 'a'
    print a == b;           @ Error: Unsupported operand types for ==: session<Cuenta> and session<Cuenta>
}
```

Una sesión se puede imprimir: `print c;` imprime `session<Cuenta>`.

### Scopes

El nombre de la sesión y el cuerpo forman un único scope, como los parámetros y el cuerpo de una función. El nombre de la falla y el bloque `rescue` forman otro (ver [Nombres y alcance en Ricol](scopes_and_names.md)). Por eso `rescue` no ve la sesión ni lo declarado en el cuerpo, y ninguno de los dos nombres existe después de la sesión:

```ricol
session conn as c {
    let c: Int = 1;  @ Error: Variable 'c' already declared in this scope
} rescue e {
    print c;         @ Error: Undefined variable 'c'
}
```

### Ejecución

1. Se evalúan la conexión y, si está, el timeout.
2. Se abre la conexión. Si falla, se pasa al paso 5.
3. Se ejecuta el cuerpo. Cada llamada remota espera la respuesta del servidor.
4. Si el cuerpo termina, o sale con `return`, `break` o `continue`, se cierra la
   conexión y la ejecución sigue según corresponda.
5. Si la apertura o una llamada fallan, se abandona el cuerpo, se cierra la
   conexión y se ejecuta `rescue`, con el nombre de la falla ligado al
   [`SessionError`](#sessionerror) que corresponde. Al terminar `rescue`, la
   ejecución sigue con la sentencia siguiente a la sesión.

Un error local en el cuerpo (por ejemplo, una división por cero) no va a `rescue`: se cierra la conexión y se aborta el programa, como con cualquier otro error local.

Tanto el cuerpo como `rescue` pueden terminar una función con `return`. En una función con tipo de retorno, una sesión cuenta como un `return` en todos los caminos solo si el cuerpo y `rescue` retornan los dos:

```ricol
func saldoRemoto() -> Int {  @ Error: Function 'saldoRemoto' does not return a value on every path
    session conn as c {
        return c.saldo();
    } rescue e {
        print "falló";
    }
}
```

### Sesiones anidadas

Una sesión puede abrirse dentro del cuerpo de otra. Cada falla se maneja en el `rescue` de la sesión a la que pertenece:

```ricol
session conn as a {
    session otraConn as b {
        a.depositar(1);     @ si esta llamada falla...
    } rescue eb {
        print "no se ejecuta";
    }
} rescue ea {
    print "...se ejecuta este rescue";
}
```

Si `a.depositar(1)` falla, se cierra `b` sin ejecutar su `rescue`, se cierra `a` y se ejecuta el `rescue` de `a`. Una falla de `b`, en cambio, se maneja en el `rescue` de `b`, que puede seguir usando `a`.

## `SessionError`

`SessionError` es un enum predefinido con las fallas de comunicación:

```ricol
enum SessionError {
    ConnectionFailed(message: String);   @ no se pudo conectar
    ConnectionLost(message: String);     @ se cortó durante una llamada
    Timeout;                             @ no llegó respuesta a tiempo
    ProtocolError(message: String);      @ interfaz incompatible o mensaje inválido
    ServerError(message: String);        @ error interno del servidor al ejecutar la función
}
```

Es un enum como cualquier otro: se puede usar como tipo, construir sus valores (`SessionError.Timeout`) y analizarlo con `match`, que tiene que cubrir todas sus variantes o tener una rama `_`.

Está declarado en el scope global, así que no se puede declarar otra cosa con ese nombre en el nivel superior del programa:

```ricol
enum SessionError { Otro; } @ Error: Enum 'SessionError' already declared in this scope
```

En un scope interno se lo puede tapar, como a cualquier nombre global, pero la falla de un `rescue` es siempre el `SessionError` predefinido.

## Tipos transportables

Los tipos de los parámetros y del retorno de los métodos de un servicio tienen que ser **transportables**, es decir, poder enviarse por la red. Son transportables los tipos primitivos, y los structs y enums cuyos campos son todos transportables. `connection<S>` y `session<S>` no lo son, y por lo tanto tampoco un struct o un enum que los contenga:

```ricol
struct Destino {
    c: connection<Cuenta>;
}

service Espejo {
    func reflejar(c: connection<Cuenta>) -> Destino;
    @ Error: Type 'connection<Cuenta>' in method 'reflejar' of service 'Espejo' is not transportable
    @ Error: Type 'Destino' in method 'reflejar' of service 'Espejo' is not transportable
}
```

Las funciones privadas de un servidor no son métodos del servicio, así que sí pueden usar estos tipos.

## Implementación

El intérprete ejecuta las sesiones a través de una interfaz de transporte, que abre conexiones y hace llamadas, y que informa cada falla como una variante de `SessionError`. El transporte que usa el intérprete hoy es el que falla siempre al abrir. Los tests del intérprete usan un transporte falso en memoria, que permite verificar el resto de la semántica: las llamadas y sus resultados, las fallas a mitad del cuerpo, el cierre de la conexión en cada forma de salir y qué `rescue` maneja cada falla en las sesiones anidadas.
