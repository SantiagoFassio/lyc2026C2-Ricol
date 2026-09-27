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
  - [gRPC](#grpc)
  - [CORBA](#corba)
  - [Java RMI](#java-rmi)
  - [Ada (Anexo E)](#ada-anexo-e)
  - [Erlang](#erlang)
  - [Jolie](#jolie)
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

Las dificultades descritas en la sección anterior no son nuevas. Desde los
primeros sistemas de RPC, distintos lenguajes y tecnologías las fueron
resolviendo de formas diferentes.

### gRPC

[gRPC](https://grpc.io) es la tecnología de RPC más usada hoy en día, y la
referencia de la que parte Ricol: un servicio es una lista de métodos con tipos
de parámetro y de retorno. El contrato se escribe en Protocol Buffers
(archivos `.proto`), y una herramienta genera, para cada lenguaje, los stubs del
cliente y la clase base que implementa el servidor.

- **Contrato**: el `.proto` es un lenguaje aparte, y el código generado es una
biblioteca más, por lo que el compilador del lenguaje de la aplicación no sabe
que esas llamadas son remotas. Además, cada método recibe un único _mensaje_ y
devuelve otro: para realizar una llamada hay que construir un `Request` y
desempaquetar un `Response`, los cuales contienen tanto los argumentos y
resultado de la ejecución del método como potencialmente un código de error.
- **Errores**: toda llamada devuelve un código de estado. Los códigos como
`FAILED_PRECONDITION` o `NOT_FOUND` nunca los genera la biblioteca, sino el
código del servidor, por lo que se usan para errores de negocio. Pero llegan
por el mismo canal que `UNAVAILABLE` o `DEADLINE_EXCEEDED`, y el cliente tiene
que distinguirlos a mano. La alternativa es modelar el error de negocio dentro
del mensaje de respuesta (con `oneof`), lo cual depende de cada diseñador.
- **Compatibilidad**: no se verifica al conectar. En cambio, cada campo de un
mensaje lleva un número, y Protocol Buffers define reglas para modificar los
mensajes sin romper a los clientes viejos (los campos desconocidos se ignoran).
Es más flexible, pero obliga a que cada campo viaje acompañado de una etiqueta.

gRPC resuelve bien el transporte y la generación de código para muchos
lenguajes, pero es justamente el ejemplo de las asperezas que Ricol busca
evitar: IDL externo, tipos generados y errores de negocio mezclados con los de
comunicación.

### CORBA

[CORBA](https://www.omg.org/spec/CORBA/) (_Common Object Request Broker Architecture_, 1991)
es un estándar de objetos remotos cuyo contrato se escribe en un IDL propio (OMG
IDL), del que se generan stubs para muchos lenguajes. Una vez generados, la
llamada remota se escribe igual que una llamada a un método local.

Su aporte más interesante está en el manejo de errores, donde es un antecedente
directo de Ricol. CORBA distingue dos jerarquías de excepciones: las
`UserException`, que se declaran en el IDL con `raises` y representan los
errores de negocio, y las `SystemException`, que cualquier operación puede
lanzar sin declararlas y representan las fallas del sistema. Además, toda
`SystemException` tiene un campo `completed` que indica qué sabe el cliente
sobre la ejecución: `COMPLETED_YES`, `COMPLETED_NO` o `COMPLETED_MAYBE`. Además,
distingue `TRANSIENT` (no se pudo establecer la comunicación) de `COMM_FAILURE`
(la comunicación se perdió después de enviar el pedido).

Ejemplo de manejo de estos dos tipos de errores en Java utilizando CORBA:

```java
try {
    Movimiento m = cuenta.retirar(30);
} catch (SaldoInsuficiente e) {              // Ejemplo de UserException
    System.out.println("saldo insuficiente: " + e.disponible);
} catch (COMM_FAILURE e) {                   // Ejemplo de SystemException
    if (e.completed == CompletionStatus.COMPLETED_MAYBE) {
        System.out.println("no se sabe si el retiro se hizo");
    }
}
```

### Java RMI

[Java RMI](https://docs.oracle.com/en/java/javase/21/docs/specs/rmi/index.html)
(_Remote Method Invocation_) elimina el IDL: el contrato es una interfaz Java
que extiende `Remote`, compartida por cliente y servidor, y que el servidor
implementa con `implements`. La llamada remota se escribe como una llamada a un
método local.

```java
public interface Cuenta extends Remote {
    Movimiento retirar(long monto) throws SaldoInsuficiente, RemoteException;
}
```

Los errores de negocio y las fallas de comunicación se separan por tipo de
excepción, y como `RemoteException` es una excepción _checked_, el compilador
obliga a manejarla. El costo es que aparece en la firma de cada método remoto y
en el manejo de cada llamada, esparciéndose por todo el código. Por otro lado,
que los argumentos sean serializables se verifica recién en tiempo de ejecución,
y la compatibilidad de versiones solo se controla parcialmente, por clase (con
el `serialVersionUID`).

Java RMI está alineado con Ricol en escribir el contrato en el propio lenguaje,
pero trata a los dos tipos de error con el mismo mecanismo (excepciones) y
distribuye el manejo de fallas por cada llamada.

### Ada (Anexo E)

El estándar de Ada incluye un anexo opcional, el
[Anexo E](https://www.adaic.org/resources/add_content/standards/22rm/html/RM-E.html)
(_Distributed Systems Annex_), que permite partir un programa en _particiones_
que se ejecutan en nodos distintos. Un paquete marcado con
`pragma Remote_Call_Interface` es la interfaz de un servicio remoto: su
especificación la ven todas las particiones y su cuerpo vive en una sola.

```ada
package Cuenta is
   pragma Remote_Call_Interface;
   function Retirar (Monto : Integer) return Resultado_Retiro;
end Cuenta;
```

Es el precedente más cercano a Ricol en varios aspectos:

- El contrato es código Ada y el compilador genera los stubs.
- El cliente llama a `Cuenta.Retirar (30)` con la misma sintaxis que una
llamada local.
- El estándar impone restricciones estáticas sobre los tipos que pueden
atravesar la red (por ejemplo, que no sean punteros locales), igual que la
noción de tipos transportables de Ricol.
- Exige que todas las particiones sean consistentes entre sí, es decir, que se
hayan construido con la misma versión de cada interfaz.

La diferencia está en las fallas: se señalan con la excepción
`System.RPC.Communication_Error`, y nada obliga a atraparla. Como la llamada es
idéntica a una local, es fácil olvidar que puede fallar, que es exactamente la
crítica realizada en el paper de Waldo.

### Erlang

En [Erlang](https://www.erlang.org/), la distribución forma parte del runtime:
un proceso puede enviar mensajes a procesos de otros nodos igual que a los
locales. El patrón más usado para RPC es el
[`gen_server`](https://www.erlang.org/doc/apps/stdlib/gen_server.html), un
proceso servidor al que se le hacen pedidos con `gen_server:call`, que espera
la respuesta con un tiempo límite.

```erlang
try gen_server:call({cuenta, 'servidor@host'}, {retirar, 30}, 2000) of
    {ok, #{saldo_final := Saldo}}       -> io:format("saldo: ~p~n", [Saldo]);
    {error, {saldo_insuficiente, Disp}} -> io:format("faltan fondos: ~p~n", [Disp])
catch
    exit:{timeout, _}       -> io:format("el servidor no respondió a tiempo~n");
    exit:{{nodedown, _}, _} -> io:format("no se pudo conectar~n")
end.
```

Los resultados de negocio son valores comunes (tuplas `{ok, ...}` y
`{error, ...}`), mientras que las fallas de distribución son _exits_ que se
atrapan aparte y que distinguen, entre otros casos, el tiempo agotado del nodo
caído. Además, un `gen_server` atiende sus pedidos de a uno, por lo que su
estado no sufre _race conditions_: es el mismo modelo que adopta el servidor de
Ricol.

Erlang coincide con Ricol en la separación de errores y en el modelo del
servidor, pero es de tipado dinámico y no tiene un contrato declarado: nada
verifica antes de ejecutar que cliente y servidor estén de acuerdo en el
formato de los mensajes.

### Jolie

[Jolie](https://jolie-lang.org/) es un lenguaje orientado a servicios,
desarrollado originalmente en la Universidad de Bolonia, en el que todo
programa es un servicio. Las interfaces y los tipos de datos se escriben en
Jolie, en archivos aparte que importan el cliente y el servidor, y la dirección
del servidor se indica como una URI, separada de la lógica:

```jolie
interface CuentaIface {
  RequestResponse:
    retirar( int )( Movimiento ) throws SaldoInsuficiente( int )
}
```

Las fallas se manejan con bloques: un `scope` agrupa varias instrucciones, e
`install` registra los manejadores de las fallas que ocurran dentro.

Es el lenguaje más parecido a Ricol en su diseño general: contrato escrito en
el lenguaje, direcciones como URI y fallas manejadas por bloque. Difiere en que
los errores de negocio y las fallas de comunicación son el mismo mecanismo
(_faults_), en que el bloque de manejo es opcional y no está ligado a la
conexión, y en que la llamada remota tiene una sintaxis propia
(`retirar@Cuenta( 30 )( mov )`).

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
    func saldo() -> Int;
    func depositar(monto: Int) -> Movimiento;
    func retirar(monto: Int) -> ResultadoRetiro;
}
```

Un `server` implementa un servicio y lo expone en una dirección:

```ricol
server Cuenta at "tcp://0.0.0.0:3825" {
    let actual: Int = 0;                             @ estado global del servidor

    func depositar(monto: Int) -> Movimiento { ... }  @ implementa un método del servicio

    func validar(monto: Int) -> Bool {                @ auxiliar privada
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

#### Contrato y servidor

En Ricol, el `service` se declara en un archivo de interfaz que importan el
cliente y el servidor, y el `server` implementa cada método con su firma
completa.

En **gRPC**, el contrato está en un `.proto`, y cada método recibe y devuelve
un mensaje en lugar de los tipos del lenguaje. El servidor implementa la clase
base que se genera a partir de ese archivo:

```proto
service Cuenta {
  rpc Retirar (RetirarRequest) returns (RetirarResponse);
}
```

En **Java RMI** y **Ada**, el contrato es código del propio lenguaje, y el
servidor lo implementa repitiendo la firma, que verifica el compilador. Java
RMI, además, exige que los argumentos sean serializables, pero lo verifica
recién durante la ejecución:

```java
public class CuentaImpl extends UnicastRemoteObject implements Cuenta {
    public Movimiento retirar(long monto) throws SaldoInsuficiente { ... }
}
```

En **Jolie**, la interfaz también se importa desde un archivo compartido, y el
servidor la expone en un `inputPort` junto con su dirección, de forma parecida
a `server Cuenta at "..."`:

```jolie
from cuenta import CuentaIface

service Servidor {
  inputPort Cuenta {
    location: "socket://0.0.0.0:3825"
    protocol: sodep
    interfaces: CuentaIface
  }
  main {
    [ retirar( monto )( mov ) { ... } ]
  }
}
```

#### Llamada y sesión

En Ricol, `connect` solo describe la conexión, `session` la abre y la cierra,
y dentro del bloque las llamadas se escriben como llamadas locales, con un
timeout por llamada. El valor `session<S>` no puede escapar de su bloque.

En **gRPC**, el cliente se crea a partir de un canal de larga vida que
reutiliza las conexiones, y cada llamada recibe explícitamente un contexto con
su tiempo límite y un mensaje de request:

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
resp, err := client.Retirar(ctx, &pb.RetirarRequest{Monto: 30})
```

En **Java RMI**, la llamada es idéntica a una local una vez obtenido el stub,
pero ese stub es un objeto común: se puede guardar en cualquier lado y usarse
en cualquier momento, sin un alcance que delimite dónde hay comunicación
remota.

```java
Cuenta c = (Cuenta) LocateRegistry.getRegistry("localhost", 3825).lookup("Cuenta");
Movimiento m = c.retirar(30);
```

**Ada** va más lejos aún: el código del cliente no menciona la conexión en
absoluto (la ubicación de cada partición se define en un archivo de
configuración aparte).

En **Jolie**, la dirección se declara en un `outputPort` (el equivalente a
`connect`), pero la llamada remota tiene una sintaxis propia, distinta de la
llamada local: `retirar@Cuenta( 30 )( mov )`.

#### Errores y `rescue`

En Ricol, los errores de negocio son variantes de un `enum` que se analizan con
`match`, y las fallas de comunicación son lo único que llega al bloque
`rescue`, el cual es obligatorio.

En **gRPC**, si el error de negocio se informa con un código de estado, llega
por el mismo `err` que las fallas de red, y el cliente tiene que distinguirlos:

```go
switch status.Code(err) {
case codes.FailedPrecondition: // error de lógica de negocio
case codes.Unavailable:        // falla de red
}
```

**Java RMI** y **CORBA** separan los dos casos por tipo de excepción, y un
mismo `try` puede agrupar varias llamadas, de forma parecida a `rescue`. Pero
los errores de negocio también son excepciones, y el manejo de las fallas es
opcional en CORBA. De CORBA, Ricol toma la clasificación de las fallas según
lo que el cliente sabe sobre la ejecución: `TRANSIENT` corresponde a
`ConnectionFailed`, y `COMM_FAILURE` con `COMPLETED_MAYBE`, a `ConnectionLost`.
En **Ada**, `System.RPC.Communication_Error` se puede atrapar en la sección
`exception` de un bloque, pero nada obliga a hacerlo.

**Erlang** es el más parecido en esta separación: los errores de negocio son
valores (`{error, ...}`) y las fallas son _exits_ que se atrapan aparte.

**Jolie** tiene la estructura de bloque más parecida a `session` / `rescue`,
pero su `scope` atrapa con el mismo mecanismo los errores de negocio y las
fallas de red:

```jolie
scope( sesion ) {
  install(
    SaldoInsuficiente => println@Console( "saldo insuficiente" )(),
    IOException       => println@Console( "falló la comunicación" )()
  );
  retirar@Cuenta( 30 )( mov )
}
```

#### Modelo del servidor

En Ricol, el servidor procesa las llamadas de a una, sobre un estado global
compartido por todos los clientes. Esta decisión de diseño resuelve el problema
de las _race conditions_ de una forma sencilla y suficiente para el alcance
esperado del lenguaje.

**Erlang** usa el mismo modelo: un `gen_server` atiende sus mensajes de a uno
en `handle_call`, y el estado solo lo modifica ese proceso:

```erlang
handle_call({retirar, Monto}, _From, Saldo) when Monto > Saldo ->
    {reply, {error, {saldo_insuficiente, Saldo}}, Saldo};
handle_call({retirar, Monto}, _From, Saldo) ->
    {reply, {ok, #{saldo_final => Saldo - Monto}}, Saldo - Monto}.
```

En **Java RMI** y **gRPC**, en cambio, las llamadas pueden ejecutarse de forma
concurrente en distintos hilos, y es responsabilidad del programador proteger
el estado compartido:

```java
public synchronized Movimiento retirar(long monto) { ... }
```

**Jolie** permite elegir: con `execution: sequential` atiende una sesión por
vez, y con `execution: concurrent` las atiende en paralelo. Además, a diferencia
de Ricol, cada sesión tiene su propio estado, y el estado compartido se accede
explícitamente con el prefijo `global`.

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
