# Servicios y servidores en Ricol

Un servicio (`service`) define un contrato: la lista de funciones que ofrece un
servidor, con los tipos de sus parámetros y de su retorno. Un servidor
(`server`) implementa un servicio y lo expone en una dirección.

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
    func reiniciar();
}

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

    func reiniciar() {
        actual = 0;
    }
}
```

## Declaración de un servicio

Un servicio se declara con la palabra reservada `service`, seguida de su nombre
y de sus métodos entre `{}`. Cada método se escribe como el encabezado de una
función (`func`, nombre, parámetros y tipo de retorno), sin cuerpo y terminado
en `;`.

Los tipos de los parámetros y del retorno pueden ser primitivos, structs o
enums, siempre que sean transportables (ver
[Tipos transportables](sessions.md#tipos-transportables)). Un método puede no
tener tipo de retorno (en el ejemplo es el caso de `reiniciar`).

Un servicio tiene que tener al menos un método, y no puede repetir métodos ni
parámetros dentro de un método:

```ricol
service Vacio {}            @ Error: Service 'Vacio' must have at least one method

service S {
    func f();
    func f(x: Int);         @ Error: Duplicate method 'f' in service 'S'
    func g(x: Int, x: String);  @ Error: Duplicate parameter 'x' in method 'g' of service 'S'
}
```

### El nombre de un servicio

El nombre de un servicio comparte los nombres con las variables, las funciones,
los structs y los enums (ver [Nombres y alcance en Ricol](scopes_and_names.md)).
Los nombres de sus métodos, en cambio, no ocupan nombres en el scope: viven
dentro del servicio, como las variantes de un enum.

Un servicio no es un tipo ni un valor:

```ricol
let c: Cuenta = 1;          @ Error: 'Cuenta' is not a type
print Cuenta;               @ Error: Cannot use service 'Cuenta' as a value
```

Su nombre solo se usa para declarar el servidor que lo implementa, y para
conectarse a él con `connect` y los tipos `connection<Cuenta>` y
`session<Cuenta>` (ver [Conexiones y sesiones en Ricol](sessions.md)).

Un servicio solo se puede declarar en el nivel superior del programa, no dentro
de un bloque, una función, una rama de un `match` o un servidor:

```ricol
{
    service S {  @ Error: Service declarations are only allowed at the top level
        func f();
    }
}
```

## Declaración de un servidor

Un servidor se declara con la palabra reservada `server`, seguida del nombre
del servicio que implementa, la palabra reservada `at`, su dirección y su
cuerpo entre `{}`.

El servicio tiene que estar declarado antes que el servidor. Al igual que un
servicio, un servidor solo se puede declarar en el nivel superior, y un programa
puede tener como mucho uno.

### Dirección

La dirección es un string con la forma `host:puerto`, con un host no vacío y un
puerto entre 1 y 65535. Las direcciones IPv6 van entre corchetes, por ejemplo
`"[::1]:3825"`. Su formato se verifica antes de ejecutar:

```ricol
server Cuenta at "tcp://0.0.0.0:3825" { ... }
@ Error: Invalid address 'tcp://0.0.0.0:3825': expected host:port
```

La dirección tiene que ser un literal: no puede ser una variable ni otra
expresión.

### Cuerpo y estado

El cuerpo del servidor es un scope, como un bloque, y puede tener cualquier
sentencia. Sus variables son el estado del servidor, que comparten todas sus
funciones. En el ejemplo, `actual` es el estado de `Cuenta`.

Las funciones del servidor ven el estado y los nombres globales del programa.
Lo declarado en el cuerpo, en cambio, no es visible después del servidor:

```ricol
server Cuenta at "0.0.0.0:3825" {
    let actual: Int = 0;
    ...
}
print actual;               @ Error: Undefined variable 'actual'
```

### Implementación del servicio

Cada método del servicio tiene que estar implementado por una función declarada
directamente en el cuerpo del servidor, con el mismo nombre, los mismos tipos de
parámetros en el mismo orden y el mismo tipo de retorno. Los nombres de los
parámetros no tienen que coincidir.

```ricol
server Cuenta at "0.0.0.0:3825" {
    func saldo() -> Float { ... }
    @ Error: Method 'saldo' of server 'Cuenta' must have signature () -> Int, got () -> Float
    ...
}
```

Si falta un método, o si está declarado dentro de un bloque del cuerpo y no
directamente en él, el error es
`Server 'Cuenta' does not implement method 'reiniciar'`.

Las funciones del servidor que no están en el servicio son privadas: solo se
pueden llamar desde el propio servidor.

Los tipos se comparan por su declaración, no por su nombre. Si el cuerpo del
servidor declara un struct con el mismo nombre que uno de afuera, es otro tipo,
y una función que lo use no implementa un método que usa el de afuera, aunque
el mensaje muestre el mismo nombre en las dos firmas.

## Ejecución

Un servidor se ejecuta cuando el programa llega a su declaración: se ejecutan
las sentencias del cuerpo, en orden, y después el programa sigue con la sentencia
siguiente. Un error en el cuerpo aborta el programa, como cualquier otro error
local.

## Formato en la red

El formato y el hash están descritos en [Formato en la red](wire.md).
