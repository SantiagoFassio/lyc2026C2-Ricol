# Structs en Ricol

Un struct es un registro con campos con nombre. Es inmutable: una vez construido, sus campos no se pueden modificar, por lo que pasarlo por copia o por referencia da lo mismo.

```ricol
struct Punto {
    x: Int;
    y: Int;
}

let a: Punto = Punto { x: 1, y: 2 };
print a;   @ Imprime todos los campos en orden.
```

Un struct es declarado como un tipo de dato, reemplaza los otros tipo de dato estandar en tipos de variables, parámetros, tipos de retorno y campos de otro struct.

## Condiciones de uso de un struct

### Declaracion de un struct

Un struct puede contener la cantidad que se quiera de campos (incluido 0). Estos campos están compuestos de nombre y tipo de dato.

```ricol
struct Punto {
    x: Int;
    y: Int;
}
```
Por ejemplo, se puede declarar el struct Punto con dos campos, x e y, ambos de tipo Int.

Los campos de un struct pueden ser otros struct, por ejemplo:

```ricol
struct Vector {
    nombre: String;
    inicio: Punto;
    final: Punto;
}
```

Esto es lo que se llama un struct anidado.

Sin embargo, un struct no admite usar su propio struct como parte de su declaracion (struct recursivo). Esto es porque en Ricol no hay nulos.

```ricol
struct Nodo {
    valor: Int;
    siguiente: Nodo; @ No es posible, devuelve "Undefined type 'Nodo'"
}
```

Para armar estructuras recursivas se usan los enums, que sí pueden usarse a sí mismos. Una variante sin el campo recursivo (como `Vacia`) es la que permite construir el primer valor (ver [Enums en Ricol](enums.md#declaración-de-un-enum)):

```ricol
enum Lista {
    Vacia;
    Nodo(valor: Int, resto: Lista);
}
```

### Construccion de un struct

Cuando se declara una variable que es de un tipo declarado anteriormente, esta debe tener todos los campos del struct, cada uno una sola vez, en cualquier orden y con su tipo exacto (Float != Int en este caso).

Igual que en una llamada, primero se chequean los valores y después se resuelve el nombre. Los campos que faltan saltan con error en orden de declaración:

```ricol
struct Dato { a: Int; b: Float; c: String; d: Bool; }
print Dato { c: "x" };
@ Error: Missing field 'a' in struct 'Dato'
@ Error: Missing field 'b' in struct 'Dato'
@ Error: Missing field 'd' in struct 'Dato'
```

### Comparaciones

Operadores de igualdad (`==`) y desigualdad (`!=`) pueden ser usados entre structs del mismo tipo. Como los booleanos, los structs no tienen orden: `<`, `<=`, `>` y `>=` son errores de tipo de dato. Tampoco se admiten operadores aritméticos (`+`, `*`, etc) ni lógicos (`and`, `or`, `not`).

### Acceso a campos de un Struct

Los campos de un struct pueden ser accedidos con el nombre de la variable y el campo necesario, de la forma `variable.campo`. Ejemplo:

```ricol
struct Punto {
    x: Int;
    y: Int;
}

let a: Punto = Punto { x: 1, y: 2 };
print a.x;   @ 1

let c: Int = a.y;
print c;     @ 2
```

Los accesos a campo:
- Se puede encadenar (`a.b.c`). En este caso `b` es otro struct (no necesariamente del mismo tipo que `a`)
- Se puede aplicar a cualquier expresión (`foo().x`).
- Se resuelve antes que cualquier operador: `-p.x` es `-(p.x)` y `p.x ** 2` es `(p.x) ** 2`.
