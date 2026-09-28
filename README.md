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
  - [Variables](#variables)
  - [if](#if)
  - [while](#while)
- [Diseño de Ricol](#diseño-de-ricol)

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

Un programa en Ricol está formado por una secuencia de sentencias, que pueden
ocupar varias líneas. Las sentencias deben finalizar en `;`, excepto las que
terminan en un bloque `{}`, como las sentencias [`if`](#if) y
[`while`](#while).

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

### Variables

```ricol
let edad: Int = 30;
edad = edad + 1;
PRINT edad;
```

Una variable se declara con la palabra reservada `let`, seguida de su nombre, su
tipo (`Int`, `Float`, `String` o `Bool`) y su valor inicial, que es
obligatorio.

El nombre de una variable debe empezar con una letra y puede contener letras,
dígitos y `_`. Se distinguen mayúsculas de minúsculas, y no puede ser una
palabra reservada.

La asignación es una expresión cuyo resultado es el valor asignado, por lo que
puede encadenarse: `a = b = 0;` asigna `0` a ambas variables.

### if

```ricol
if (10 % 2 == 0) {
    PRINT "10 es par";
} else {
    PRINT "10 es impar";
}
```

La condición debe una expresión cuyo resultado sea de tipo `Bool`. Va siempre
entre `()` y cada rama es un bloque delimitado por `{}`, aunque tenga una sola
sentencia.

Un bloque `{}` también puede usarse por sí solo, fuera de un `if`, para agrupar
sentencias. A diferencia del resto de las sentencias, ni el `if` ni los bloques
terminan en `;`.

### while

```ricol
while (True) {
    PRINT "se imprime una sola vez";
    break;
}
```

El cuerpo se ejecuta mientras la condición sea verdadera. Al igual que en el
`if`, la condición debe ser una expresión de tipo `Bool` y va siempre entre
`()`, y el cuerpo es un bloque delimitado por `{}`. El `while` tampoco termina
en `;`.

Dentro del cuerpo pueden usarse dos sentencias, que sí terminan en `;`:

- `break;` termina el `while` y la ejecución sigue con la sentencia siguiente.
- `continue;` saltea el resto del cuerpo y vuelve a evaluar la condición.

Ambas afectan solo al `while` más cercano que las contiene. Usarlas fuera de un
`while` es un error, que se detecta antes de ejecutar el programa.

## Diseño de Ricol

En el siguiente documento se detallan varios puntos relacionados con el diseño
de Ricol, como la motivación del lenguaje, lenguajes y tecnologías precedentes y
la sintaxis propuesta para el manejo de Remote Procedure Calls:

[Diseño de Ricol](docs/design.md).
