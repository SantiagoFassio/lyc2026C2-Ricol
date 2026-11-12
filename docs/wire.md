# Formato en la red en Ricol

Cuando un cliente llama a un método de un servicio, los argumentos y el
resultado viajan por la red como bytes. El paquete `wire` define cómo se
serializa cada valor de Ricol en bytes, y calcula el hash de la interfaz de cada
servicio, que permite verificar que cliente y servidor usan el mismo contrato
(ver [Diseño de Ricol](design.md#formato-en-la-red-y-hash-de-la-interfaz)).

## Formato de los valores

Como cliente y servidor conocen la interfaz, los valores viajan sin nombres ni
etiquetas de tipo. Todos los números se escriben en big-endian, sin importar el
procesador:

| Tipo | Serialización |
| --- | --- |
| `Int` | 8 bytes, entero con signo en complemento a dos. |
| `Float` | 8 bytes, IEEE 754 de doble precisión. Los infinitos y `NaN` se transmiten. |
| `Bool` | 1 byte: `0x00` es `False` y `0x01` es `True`. |
| `String` | 4 bytes con el largo en bytes (sin signo), seguidos del texto en UTF-8. |
| struct | Sus campos en orden de declaración, sin encabezado. |
| enum | 4 bytes con el índice de la variante (sin signo, en orden de declaración), seguidos de los campos de esa variante. |
| `Void` | Nada. |

Los argumentos de un método son sus parámetros uno detrás de otro, sin cantidad
ni separadores. Un método sin parámetros y sin retorno, como `reiniciar`, no
envía ningún byte de valores.

Ejemplos con el servicio `Cuenta` de [Servicios y servidores](services.md):

| Valor | Bytes |
| --- | --- |
| argumentos de `depositar(30)` | `00 00 00 00 00 00 00 1E` |
| `Movimiento { monto: 30, saldoFinal: 130 }` | `00 00 00 00 00 00 00 1E` `00 00 00 00 00 00 00 82` |
| `ResultadoRetiro.SaldoInsuficiente(100)` | `00 00 00 01` `00 00 00 00 00 00 00 64` |
| `ResultadoRetiro.Ok(Movimiento { monto: -30, saldoFinal: 70 })` | `00 00 00 00` `FF FF FF FF FF FF FF E2` `00 00 00 00 00 00 00 46` |
| `"hola"` | `00 00 00 04` `68 6F 6C 61` |

### Datos inválidos

Los bytes que llegan por la red pueden estar mal formados, por error o a
propósito. Deserializarlos nunca hace fallar al intérprete: devuelve un error
`ErrInvalidData` cuando:

- los datos se terminan en medio de un valor;
- un `Bool` no es `0x00` ni `0x01`;
- un `String` no es UTF-8 válido;
- el índice de una variante no corresponde a ninguna variante del enum;
- el valor tiene demasiados niveles de anidamiento (ver
  [Profundidad](#profundidad)).

Si falla la lectura por otro motivo, por ejemplo porque se cortó la conexión, se
devuelve ese error y no `ErrInvalidData`, para poder distinguir los dos casos.

El texto de un `String` se lee de a bloques de 64 KB, y no reservando de una vez
la memoria que indica su largo. Si no, un mensaje de pocos bytes que dijera
tener un string de 4 GB haría reservar 4 GB antes de descubrir que los datos no
están.

Al serializar también se verifica que un `String` sea UTF-8 válido y que su
largo entre en 4 bytes, antes de escribir nada.

### Profundidad

Cada struct y cada valor de un enum suman un nivel de anidamiento. Por ejemplo,
con un enum recursivo como:

```ricol
enum Lista {
    Vacia;
    Nodo(valor: Int, resto: Lista);
}
```

una lista de `n` elementos tiene `n + 1` niveles: `n` variantes `Nodo` y la
variante `Vacia` del final.

El máximo es de 100000 niveles. Deserializar un valor más profundo da
`ErrInvalidData`, y serializarlo da `ErrTooDeep`. El límite evita que un mensaje
malicioso agote el stack del servidor, y es lo bastante alto para no rechazar
listas razonables. Aplicarlo también al serializar hace que el cliente falle
antes de enviar algo que el servidor va a rechazar.

## Forma canónica y hash

El hash de la interfaz es el SHA-256 de una representación en texto del
servicio y de los tipos que usa, su **forma canónica**. Para el servicio
`Cuenta`:

```plaintext
ricol-interface-v1
service Cuenta
method saldo() -> Int
method depositar(Int) -> Movimiento
method retirar(Int) -> ResultadoRetiro
method reiniciar() -> Void
struct Movimiento { monto: Int; saldoFinal: Int; }
enum ResultadoRetiro { Ok(mov: Movimiento); SaldoInsuficiente(disponible: Int); }
```

- La primera línea es la versión del formato. Si alguna vez cambia la
  serialización de los valores, se cambia la versión y cambian todos los hashes,
  aunque las interfaces sean iguales.
- Los métodos van en orden de declaración, con los tipos de sus parámetros pero
  sin sus nombres.
- Después vienen los structs y enums que usa el servicio, cada uno una sola vez,
  en el orden en que aparecen al recorrer los métodos (primero los parámetros y
  después el retorno). El recorrido es en profundidad: cuando aparece un tipo
  nuevo, se escriben enseguida los tipos de sus campos. Si un método recibe
  `(Dibujo, Segmento)`, y los dos usan `Punto`, el orden es `Dibujo`, `Punto`,
  `Segmento`.
- Dentro de los métodos y de los campos, un struct o un enum se escribe con su
  nombre. Eso también resuelve los tipos recursivos:
  `enum Lista { Vacia; Nodo(valor: Int, resto: Lista); }`.
- Cada campo de un struct y cada variante de un enum terminan en `;`. Un struct
  sin campos queda `struct Nada { }`.
- Las líneas se separan con `\n`, sin `\n` después de la última.

Usar los nombres para referirse a los tipos es seguro: todos los tipos que usa
un servicio son globales, porque un servicio solo se declara en el nivel
superior, y en el scope global no puede haber dos declaraciones con el mismo
nombre. Tampoco se confunden con los tipos primitivos, porque `Int`, `String` y
los demás son palabras reservadas.

### Qué cambia el hash

| Cambio en la interfaz | ¿Cambia el hash? |
| --- | --- |
| Renombrar un parámetro | No |
| Agregar, quitar o reordenar métodos, parámetros, campos o variantes | Sí |
| Cambiar un tipo (por ejemplo, `Int` por `Float`) | Sí |
| Renombrar el servicio, un método, un struct, un campo, un enum o una variante | Sí |
| Cambiar un tipo que el servicio no usa | No |
| Cambiar las funciones privadas o el estado del `server` | No |

Los nombres de los parámetros quedan afuera porque son posicionales. Los demás
nombres sí entran, aunque no viajen en los mensajes. Por ejemplo,
`Movimiento { monto: Int; saldoFinal: Int; }` y
`Movimiento { saldoFinal: Int; monto: Int; }` se serializan igual, porque los
dos campos son `Int`, pero un lado leería el saldo donde el otro escribió el
monto. Como el formato no lleva nombres, el hash es lo único que detecta ese
cambio.
