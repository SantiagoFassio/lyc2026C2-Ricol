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
  - [Operaciones booleanas](#operaciones-booleanas)
  - [Variables](#variables)
  - [if](#if)
  - [while](#while)
  - [Funciones](#funciones)
  - [Structs](#structs)
  - [Enums](#enums)
  - [match](#match)
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

Los tests de integración son programas de Ricol, en `integration/`, que se
verifican a sí mismos. Se ejecutan desde la raíz del repositorio con
`make integration`.

Además, con `make bench` se ejecuta un benchmark que compara la ejecución de
Ricol con otros lenguajes (C, Go y Python). En
[Resultados de los benchmarks](benchmarks/results/results.md) se puede ver una
corrida de ejemplo con algunas conclusiones.

## Sintaxis de Ricol

La gramática formal de Ricol se encuentra en
[Backus-Naur Form de Ricol](docs/bnf.md).

### Hola mundo

```ricol
print "Hola mundo";
```

La palabra reservada `print` permite imprimir valores por pantalla. Puede
imprimir valores de cualquier tipo, no solo strings. No agrega salto de línea al
final.

### Sentencias

Un programa en Ricol está formado por una secuencia de sentencias, que pueden
ocupar varias líneas. Las sentencias deben finalizar en `;`, excepto las que
terminan en un bloque `{}`, como las sentencias [`if`](#if),
[`while`](#while) y [`match`](#match) y las declaraciones de
[funciones](#funciones), de [structs](#structs) y de [enums](#enums).

### Tipos de datos

Ricol es un lenguaje de tipado **estático** y **fuerte**. Posee los siguientes
tipos de datos primitivos:

- **Int**: números enteros de 64 bits, con signo. Ejemplo: `3`.
- **Float**: números de punto flotante de 64 bits según lo establecido por el
estándar IEEE 754, con la excepción de que dividir por cero es un error.
Ejemplo: `41.3`.
- **String**: cadenas de carácteres. Ejemplo: `"Ricol"`.
- **Bool**: valores de verdad, `True` o `False`.

Además, se pueden declarar tipos compuestos por campos con nombre, los
[structs](#structs), y tipos cuyos valores son una entre varias variantes, los
[enums](#enums).

Que el tipado sea **estático** significa que los tipos de todas las expresiones
se verifican antes de ejecutar el programa. Si hay algún error de tipos, se
informan todos juntos y no se ejecuta ninguna sentencia. Por ejemplo, el
siguiente programa no imprime nada:

```ricol
print "Hola";
"a" - "b";      @ Error de tipos: los strings no se pueden restar
```

Que sea **fuerte** significa que no hay conversiones implícitas entre tipos,
excepto de `Int` a `Float` en las operaciones aritméticas.

Más información en la documentación de los tipos:

- [Números](docs/numbers.md)
- [String](docs/strings.md)
- [Booleanos](docs/bools.md)
- [Structs](docs/structs.md)
- [Enums](docs/enums.md)

### Comentarios

```ricol
@ Línea con un comentario

print "Ricol"; @ Comentario inline
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
print "Hola, " + "mundo";   @ Imprime Hola, mundo
```

El operador `+` concatena dos strings. Es el único operador que admiten los
strings, y ambos operandos deben ser strings.

Más información en [Strings en Ricol](docs/strings.md).

### Operaciones booleanas

```ricol
2 < 3;                      @ True
"hola" == "Hola";           @ False
1 == 1.0;                   @ True
not 10 % 2 == 0;            @ False
1 < 2 and 2 < 3;            @ True
True or False and False;    @ True: `and` se resuelve antes que `or`
```

Los operadores de comparación (`<`, `<=`, `>`, `>=`) y de igualdad (`==`,
`!=`) dan como resultado un `Bool`. Ambos operandos deben ser del mismo tipo,
con la excepción de que un `Int` puede compararse con un `Float` (ambos son del tipo `Number`): `1 == 1.0` es
`True`, pero `1 == "1"` es un error de tipos. Los strings se ordenan según el
código Unicode de sus caracteres, por lo que las mayúsculas van antes que las
minúsculas (`"Z" < "a"` es `True`). Los `Bool` pueden compararse por igualdad, pero no tienen orden.

Los operadores lógicos `not`, `and` y `or` solo admiten operandos de tipo
`Bool`. `and` y `or` evalúan el operando derecho solo si hace falta
(_cortocircuito_): en `False and f()`, la función `f` no se ejecuta.

Primero se resuelven las operaciones aritméticas, y después, en este orden, las
comparaciones, la igualdad, `not`, `and` y `or`. Por eso `not 10 % 2 == 0` es
`not ((10 % 2) == 0)`, y `a or b and c` es `a or (b and c)`. Se pueden usar
`()` para agrupar.

Más información en [Booleanos en Ricol](docs/bools.md).

### Variables

```ricol
let edad: Int = 30;
edad = edad + 1;
print edad;
```

Una variable se declara con la palabra reservada `let`, seguida de su nombre, su
tipo (`Int`, `Float`, `String`, `Bool` o el nombre de un [struct](#structs)) y
su valor inicial, que es obligatorio.

El nombre de una variable debe empezar con una letra y puede contener letras,
dígitos y `_`. Se distinguen mayúsculas de minúsculas, y no puede ser una
palabra reservada.

La asignación es una expresión cuyo resultado es el valor asignado, por lo que
puede encadenarse: `a = b = 0;` asigna `0` a ambas variables.

### if

```ricol
if (10 % 2 == 0) {
    print "10 es par";
} else {
    print "10 es impar";
}
```

La condición debe ser una expresión cuyo resultado sea de tipo `Bool`. Va siempre
entre `()` y cada rama es un bloque delimitado por `{}`, aunque tenga una sola
sentencia.

Un bloque `{}` también puede usarse por sí solo, fuera de un `if`, para agrupar
sentencias. A diferencia del resto de las sentencias, ni el `if` ni los bloques
terminan en `;`.

### while

```ricol
while (True) {
    print "se imprime una sola vez";
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

### Funciones

```ricol
func sumar(a: Int, b: Int) -> Int {
    return a + b;
}

func saludar(nombre: String) {
    print "hola, " + nombre;
}

print sumar(2, 3);
saludar("Ricol");
```

Una función se declara con la palabra reservada `func`, seguida de su nombre,
sus parámetros entre `()` separados por `,` y, opcionalmente, su tipo de
retorno después de `->`. Cada parámetro lleva su tipo, igual que una variable.
Una función sin tipo de retorno no devuelve ningún valor.

Los argumentos de las funciones se evalúan de izquierda a derecha.

Las funciones no son valores: no pueden guardarse en variables, pasarse como
argumento ni imprimirse.

Las variables, los parámetros, las funciones, los structs y los enums comparten
los nombres: en un mismo scope no puede haber dos declaraciones que se llamen igual,
y en un scope interno una puede tapar a otra. Más información en
[Nombres y alcance en Ricol](docs/scopes_and_names.md).

Hay un límite de 10000 llamadas anidadas. Superarlo, por ejemplo con una
recursión infinita, es un error de ejecución.

### Structs

```ricol
struct Punto {
    x: Int;
    y: Int;
}

let p: Punto = Punto { y: 2, x: 1 };
print p;            @ Imprime Punto { x: 1, y: 2 }
print p.x + p.y;    @ Imprime 3
```

Un struct es un tipo de dato compuesto por campos con nombre. Se declara con la
palabra reservada `struct`, seguida de su nombre y de sus campos entre `{}`.
Cada campo lleva su tipo, igual que una variable, y termina en `;`. Un campo
puede ser de otro struct declarado antes, pero un struct no puede contenerse a
sí mismo.

El nombre de un struct se puede usar como tipo de variables, parámetros, tipos
de retorno y campos de otros structs.

Un valor se construye con el nombre del struct y el valor de cada campo entre
`{}`, separados por `,`. Hay que indicar todos los campos, una sola vez cada
uno, en cualquier orden y con su tipo exacto. A un campo se accede con `.`, por
ejemplo `p.x`.

Los structs son inmutables: `p.x = 3;` es un error. Para cambiar un valor hay
que construir uno nuevo, por ejemplo `p = Punto { x: 3, y: p.y };`.

Dos structs del mismo tipo se pueden comparar con `==` y `!=`, que comparan
campo por campo. Los structs no tienen orden ni admiten otros operadores.

Un struct se puede declarar en cualquier scope, y es visible desde la sentencia
siguiente a su declaración.

Más información en [Structs en Ricol](docs/structs.md). El recorrido de un
struct por las fases del intérprete se describe en
[Recorrido de la declaración de un struct](docs/struct_tecnico.md).

### Enums

```ricol
enum Forma {
    Punto;
    Circulo(radio: Float);
    Rectangulo(ancho: Float, alto: Float);
}

let f: Forma = Forma.Rectangulo(2.0, 3.5);
print f;                            @ Imprime Forma.Rectangulo(2, 3.5)
print f == Forma.Punto;             @ Imprime False
```

Un enum es un tipo de dato cuyos valores son una entre varias variantes. Se
declara con la palabra reservada `enum`, seguida de su nombre y de sus variantes
entre `{}`. Cada variante termina en `;` y puede tener campos entre `()`.

Un valor se construye con el nombre del enum, `.` y el nombre de la variante.

Los enums son inmutables. Para usar los campos de un valor se usa
[`match`](#match).

Más información en [Enums en Ricol](docs/enums.md).

### match

```ricol
func area(f: Forma) -> Float {
    match (f) {
        Punto => return 0.0;
        Circulo(r) => return 3.14 * r * r;
        Rectangulo(ancho, alto) => return ancho * alto;
    }
}

print area(Forma.Circulo(1.0));     @ Imprime 3.14
```

La sentencia `match` ejecuta una rama distinta según la variante de un valor de
un enum. La expresión va entre `()`, como en `if` y `while`, y cada rama tiene un
patrón, `=>` y una sentencia, que puede ser un bloque. Se ejecuta la primera rama
cuyo patrón coincide con el valor.

Un patrón es el nombre de una variante, con un nombre entre `()` por cada uno de
sus campos, que se ligan a los valores de los campos en orden. `_` en lugar de
un nombre ignora ese campo, y `_` solo es la rama por defecto, que coincide con
cualquier valor.

Un `match` tiene que cubrir todas las variantes del enum o tener una rama `_`, y
no puede tener ramas que nunca se ejecuten.

Más información en [Enums en Ricol](docs/enums.md#match).

## Diseño de Ricol

En el siguiente documento se detallan varios puntos relacionados con el diseño
de Ricol, como la motivación del lenguaje, lenguajes y tecnologías precedentes y
la sintaxis propuesta para el manejo de Remote Procedure Calls:

[Diseño de Ricol](docs/design.md).
