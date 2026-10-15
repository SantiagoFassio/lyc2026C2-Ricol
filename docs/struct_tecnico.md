# Recorrido de la declaracion de un struct

## Scanner

- `struct` es una palabra reservada (token `STRUCT`), al igual que `func`, por ejemplo.
- El `.` genera el token `DOT`, que se usa para acceder a los campos.
- Los nombres de los structs y de sus campos son tokens de tipo `IDENTIFIER`, igual que los de las variables.

Por ejemplo, `print a.x;` genera:

```
PRINT<print>
IDENTIFIER<a>
DOT<.>
IDENTIFIER<x>
SEMICOLON<;>
```

Se cuida que DOT<.> (el de acceder a un campo de un struct) no cause conflicto con el punto de un literal Numerico.

## Parser

El parser define tres estructuras posibles de parte a un struct: declaración, el literal (construir un valor de tipo struct) y el acceso a campo.

- `struct Punto { ... }` genera una declaración de struct con su nombre y sus campos. Cada campo termina en `;`. Esto genera un `StructDeclarationStatement`.
- Un `IDENTIFIER` en un lugar donde va un tipo de dato (Es decir, en vez de `Int`, se ve `Punto`) genera un `NamedType`. El parser no sabe que nombres son tipos ya declarados y que nombres no. `let x: abcd = 1;` se parsea sin errores. El Typechecker se encarga de detectar que `abcd` no es un struct declarado anteriormente.
- Un `IDENTIFIER` seguido de `{` en una expresión es un literal de un struct. Los campos se separan con comas (`,`).
- `.campo` después de una expresión es un acceso a un campo. 
    - Se puede encadenar (`a.b.c`)
    - Se puede aplicar a cualquier expresión (`f().x`, `Punto { x: 1, y: 2 }.x`)
    - Se resuelve antes que cualquier operador: `-p.x` es `-(p.x)` y `p.x ** 2` es `(p.x) ** 2`.
- `p.x = 1;` no es posible, ya que los structs son inmutables.

## Typechecker

### Declaración

En la declaración de un `struct` se resuelven primero los tipos de sus campos y después declara el struct en el scope actual (es decir, resuelve los valores internos antes de hacer el completo, al igual que cuando se declaran variables).

El struct comparte el espacio de nombres con las variables y las funciones: no puede llamarse igual que otra declaración del mismo scope, pero se puede tapar a una de un scope externo. Esto aplica de Structs a Funciones, Structs a Structs, viceversa, etc (ver scopes_and_names.md)

Como los campos se resuelven antes de declarar el struct, un struct no puede contenerse a sí mismo.

### Nombres de tipo de dato

Cada `NamedType` (de variables, parámetros, tipos de retorno y campos) se busca en los scopes y se resuelve al `*StructType` de la declaración que encuentra.

Si el nombre no existe, el error es `Undefined type 'X'`, y si es una variable o una función, `'x' is not a type`.

## Intérprete

Una declaración `struct` se registra en el ambiente actual, en su propia tabla (como las variables y funciones), para que los literales conozcan el orden de los campos. Si el struct se declara dentro de una función, se registra cada vez que se llame a la funcion (como las variables).

Un literal busca su declaración con la distancia que calculó el typechecker (igual a llamar a una funcion). Se evalúan los valores de izquierda a derecha, en el orden en que se escribieron en el codigo, y arma el `Struct` con los campos en orden de declaración del struct.

Como los structs son inmutables, asignar un struct a otra variable o pasarlo a una función no comparte nada que se pueda modificar: reasignar una de las variables no cambia a la otra.

