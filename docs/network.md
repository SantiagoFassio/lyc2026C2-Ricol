# Runtime de red en Ricol

El runtime de red es lo que hace que un servidor atienda llamadas y que una sesión se comunique con él. El servidor escucha conexiones TCP en su dirección, y cada sesión del cliente abre una conexión, verifica con un handshake que los dos usan la misma interfaz y envía las llamadas. Cada falla de comunicación se clasifica en una variante de [`SessionError`](sessions.md#sessionerror).

Los valores viajan con el formato de [Formato en la red](wire.md). Este documento describe cómo se ejecutan el servidor y el cliente, qué falla corresponde a cada situación y el protocolo con el que se comunican.

## Ejecutar un servidor y un cliente

En `examples/sessions/` hay un servidor y un cliente del servicio `Account`. Desde `src/`, en una terminal se levanta el servidor:

```sh
go run . ../examples/sessions/server.ric
```

Y en otra se ejecuta el cliente:

```sh
go run . ../examples/sessions/client.ric
```

```plaintext
withdrawal ok, balance: 70
insufficient funds, available: 70
final balance: 70
```

El servidor sigue escuchando hasta que se lo interrumpa con Ctrl+C. Como su estado se mantiene entre sesiones, si se vuelve a ejecutar el cliente el saldo final es 140. Si no hay ningún servidor escuchando, el cliente imprime:

```plaintext
could not connect: cannot connect to Account at localhost:3825: connection refused
```

## El servidor

Un programa que declara un `server` primero ejecuta todas sus sentencias, en orden, incluido el cuerpo del servidor. Al terminar, empieza a escuchar en la dirección del servidor, hasta que se lo interrumpa.

Si no puede escuchar, por ejemplo porque otro programa ya usa el puerto, es un error local, con la posición de la dirección:

```plaintext
[servidor.ric, line 3, column 18] Cannot listen on 0.0.0.0:3825: address already in use
```

### Llamadas secuenciales

El servidor ejecuta las llamadas de a una en orden de llegada, sin importar de qué cliente vengan. Si dos clientes tienen sesiones abiertas al mismo tiempo, sus llamadas se intercalan, pero nunca se ejecutan dos funciones del servidor a la vez. De esta forma, se garantiza que no hayan race conditions en el estado del servidor, y cada llamada ve los cambios de todas las anteriores.

La contra de este diseño es que las llamadas lentas demoran a las siguientes llamadas en la cola: mientras se ejecuta, las llamadas de los otros clientes esperan su turno, y pueden agotar su timeout.

Nota: Una mejora de diseño podria ser implementar round robin para las llamadas.

El handshake y la lectura de los mensajes no esperan su turno: se hacen para cada conexión por separado, mientras se ejecuta otra llamada. Solo la ejecución de las funciones es de a una por vez.

### Errores en un método

Si un método tiene un error local, el cliente recibe un `ServerError` con el mensaje del error, y el servidor lo escribe en la salida de errores:

```plaintext
Method 'retirar' of server 'Cuenta' failed: [servidor.ric, line 18, column 25] Cannot divide by zero: 10 // 0
```

El servidor sigue atendiendo llamadas. El estado queda como lo dejó la función hasta el error: los cambios anteriores al error no se deshacen.

## El cliente

Cada sesión abre su propia conexión TCP al empezar y la cierra al terminar, sea cual sea la forma en que termina.

El timeout de la sesión (ver [Timeout](sessions.md#timeout)) se aplica de dos formas:

- **A la apertura:** un único límite para conectarse y hacer el handshake.
- **A cada llamada:** un límite nuevo por llamada, desde que se empieza a enviar el pedido hasta que llega la respuesta completa.

Antes de enviar una llamada se serializan todos sus argumentos. Si un argumento no se puede serializar (por ejemplo, un valor con demasiados niveles de anidamiento, ver [Profundidad](wire.md#profundidad)), es un error local en la posición de la llamada, y no se envía nada:

```plaintext
[cliente.ric, line 7, column 7] Cannot send the arguments of 'sumar': value too deep: more than 100000 levels of nesting
```

## Clasificación de las fallas

Las fallas se clasifican según lo que el cliente sabe sobre lo que pasó en el
servidor:

| Situación | Falla | ¿Se ejecutó el método? |
| --- | --- | --- |
| No se pudo conectar: la conexión fue rechazada, el host no existe, el servidor cerró la conexión o se agotó el tiempo durante la apertura | `ConnectionFailed` | No |
| El servidor no es un programa Ricol, usa otra versión del protocolo o implementa otra interfaz | `ProtocolError` | No |
| No llegó la respuesta a tiempo | `Timeout` | Si/No |
| Se cortó la conexión durante una llamada | `ConnectionLost` | Si/No |
| El método tuvo un error local en el servidor | `ServerError` | Sí, hasta el error |
| El servidor rechazó el pedido o la respuesta es inválida | `ProtocolError` | Depende |

Un `ProtocolError` durante una llamada puede venir del servidor, que rechaza un pedido mal formado antes de ejecutarlo, o del cliente, que recibe una respuesta que no puede leer, cuando el método ya se ejecutó. Con un programa Ricol de cada lado, que comparten la interfaz, no debería ocurrir nunca.

Un timeout que se agota durante la apertura es `ConnectionFailed` y no `Timeout`. Así, `ConnectionFailed` siempre significa que no se ejecutó nada, y `Timeout` que no se sabe.

Los mensajes de las fallas de la apertura empiezan con el servicio y la dirección, seguidos de la causa:

```plaintext
cannot connect to Cuenta at localhost:3825: connection refused
cannot connect to Cuenta at localhost:3825: timed out after 2s
cannot connect to Cuenta at localhost:3825: the server is not a Ricol program
cannot connect to Cuenta at localhost:3825: interface mismatch with the server, which implements service 'Contador'
```

Y los de una conexión cortada nombran el método:

```plaintext
connection to Cuenta at localhost:3825 lost during 'depositar': the connection was closed
```

### At most once

Si una llamada devuelve un resultado, el método se ejecutó exactamente una vez. Si falla con `Timeout` o `ConnectionLost`, puede haberse ejecutado una vez o ninguna. Es la semántica "at most once" que describen Birrell y Nelson (ver [Diseño de Ricol](design.md#motivación-de-ricol)).

El método puede ejecutarse incluso después de la falla. Si una llamada agota su timeout mientras espera su turno detrás de una llamada lenta, el servidor la ejecuta igual cuando le llega el turno, aunque el cliente ya no esté esperando la respuesta.

## Protocolo

Cliente y servidor se comunican por TCP. Todos los números se escriben en big-endian, como en [Formato en la red](wire.md).

### Handshake

Al conectarse, el cliente envía un saludo de 38 bytes:

| Bytes | Contenido |
| --- | --- |
| 5 | `RICOL` (`52 49 43 4F 4C`) |
| 1 | Versión del protocolo: `01` |
| 32 | Hash de la interfaz del servicio (ver [Forma canónica y hash](wire.md#forma-canónica-y-hash)) |

El servidor responde con los mismos 6 primeros bytes (`RICOL` y la versión), seguidos de un frame de respuesta (ver [Respuestas](#respuestas)):

- Si el hash coincide, la respuesta es `OK` sin contenido, y la sesión queda abierta: `52 49 43 4F 4C 01` `00 00 00 01` `00`.
- Si el hash no coincide, la respuesta es `ProtocolError`, con un mensaje que nombra el servicio que implementa el servidor, y el servidor cierra la conexión. Lo mismo pasa si la versión del protocolo es otra.

Si los primeros 5 bytes no son `RICOL`, el otro lado no es un programa Ricol: el servidor cierra la conexión sin responder, y el cliente falla con `ProtocolError`.

Como el saludo tiene un tamaño fijo, el servidor nunca hace una lectura de largo arbitrario antes de saber que el cliente es un programa Ricol con la misma interfaz.

### Frames

Después del handshake, cada mensaje es un frame: 4 bytes con el largo del contenido, seguidos del contenido.

Los frames permiten distinguir una conexión que se cortó a mitad de un mensaje (`ConnectionLost`) de un mensaje completo pero mal formado (`ProtocolError`). El contenido se lee de a bloques de 64 KB, igual que los strings, así que un largo falso no hace reservar memoria de más.

### Llamadas

Una llamada es un frame con el índice del método (4 bytes, sin signo, en orden de declaración en el servicio), seguido de los argumentos. Con el servicio `Cuenta` de [Servicios y servidores](services.md):

| Llamada | Bytes |
| --- | --- |
| `c.saldo()` | `00 00 00 04` `00 00 00 00` |
| `c.depositar(30)` | `00 00 00 0C` `00 00 00 01` `00 00 00 00 00 00 00 1E` |

Si el índice no corresponde a ningún método, si los argumentos no se pueden leer, o si sobran bytes después de ellos, el servidor responde `ProtocolError` sin ejecutar nada y cierra la conexión.

### Respuestas

Una respuesta es un frame con un byte de estado seguido de su contenido:

| Estado | Contenido |
| --- | --- |
| `00`: OK | El resultado del método |
| `01`: `ServerError` | El mensaje del error, en UTF-8 |
| `02`: `ProtocolError` | El mensaje del error, en UTF-8 |

| Respuesta | Bytes |
| --- | --- |
| `Movimiento { monto: 30, saldoFinal: 130 }` | `00 00 00 11` `00` `00 00 00 00 00 00 00 1E` `00 00 00 00 00 00 00 82` |
| Resultado de `c.reiniciar()` | `00 00 00 01` `00` |
| Error `division by zero` | `00 00 00 11` `01` y los 16 bytes del texto |

Los mensajes de error se cortan a 64 KB. Si el cliente recibe un estado desconocido, un resultado que no puede leer, bytes de más después del resultado o un mensaje que no es UTF-8 válido, la falla es `ProtocolError`.

## Implementación

El runtime de red está en el paquete `network`, que no depende del intérprete:

- `protocol.go`: el saludo, los frames, las llamadas y las respuestas.
- `client.go`: `Dial` (la apertura con el handshake) y `Call`, que clasifican
  cada falla.
- `server.go`: `Listen` y `Serve`. Cada conexión se atiende en un goroutine
  propio, que hace el handshake y lee los pedidos, y le pasa cada llamada a
  `Serve` por un canal. `Serve` ejecuta las llamadas de a una, en el goroutine
  del intérprete.

El intérprete ejecuta las sesiones a través de una interfaz de transporte (ver [Implementación](sessions.md#implementación)), y el transporte de red la implementa con el paquete `network`. Después de la verificación de tipos, el intérprete arma la interfaz de cada servicio con el paquete `wire`, ejecuta el programa y, si declaró un servidor, escucha.
