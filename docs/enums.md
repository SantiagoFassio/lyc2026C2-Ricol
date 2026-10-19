# Enums en Ricol

Un enum es un tipo de dato cuyos valores son una entre varias variantes. Cada
variante puede tener campos propios. Como los structs, los enums son inmutables.

```ricol
enum Forma {
    Punto;
    Circulo(radio: Float);
    Rectangulo(ancho: Float, alto: Float);
}

let f: Forma = Forma.Rectangulo(2.0, 3.5);
print f;   @ Forma.Rectangulo(2, 3.5)
```

El nombre de un enum se usa como tipo, igual que el de un struct.

## Declaración de un enum

Cada variante termina en `;`. Si tiene campos, van entre `()`, separados por
`,`, cada uno con su tipo. Una variante sin campos se escribe sin paréntesis.
Un enum tiene que tener al menos una variante, y no puede repetir variantes ni
campos dentro de una variante.

A diferencia de un struct, un enum puede usarse a sí mismo en sus campos, porque
es visible desde su propia declaración. La variante sin el campo recursivo es la
que permite construir el primer valor:

```ricol
enum Lista {
    Vacia;
    Nodo(valor: Int, resto: Lista);
}

let numeros: Lista = Lista.Nodo(1, Lista.Nodo(2, Lista.Vacia));
```

## Construcción de un valor

Un valor se construye con el nombre del enum, `.` y la variante. Los valores de
los campos van entre `()`, en el orden en que se declararon y con su tipo
exacto:

```ricol
print Forma.Punto;        @ Forma.Punto
print Forma.Circulo;      @ Error: Variant 'Circulo' of enum 'Forma' expects 1 argument, got 0
print Forma.Punto();      @ Error: Variant 'Punto' of enum 'Forma' has no fields
print Forma.Circulo(1);   @ Error: Field 'radio' of variant 'Circulo' of enum 'Forma' must be Float, got Int
```

Los campos de un enum no se acceden con `.`: para usarlos hay que usar `match`.

## Comparaciones

`==` y `!=` se pueden usar entre valores del mismo enum: son iguales si tienen
la misma variante y los mismos campos. Como los structs, los enums no tienen
orden ni admiten otros operadores.

## `match`

`match` ejecuta la primera rama cuyo patrón coincide con la variante del valor:

```ricol
func area(f: Forma) -> Float {
    match (f) {
        Punto => return 0.0;
        Circulo(r) => return 3.14 * r * r;
        Rectangulo(ancho, alto) => return ancho * alto;
    }
}
```

La expresión va entre paréntesis, como en `if` y `while`. Cada rama tiene un
patrón, `=>` y una sentencia, que puede ser un bloque.

- `Variante(a, b)` liga los campos, en orden, a los nombres `a` y `b`. Tiene que
haber un nombre por campo.
- `_` en lugar de un nombre ignora ese campo.
- `_` solo es la rama por defecto.

No hay patrones anidados ni con literales.

Un `match` tiene que cubrir todas las variantes o tener una rama `_`, y no puede
tener ramas que nunca se ejecuten:

```ricol
match (f) {
    Punto => print "punto";
    Circulo(r) => print r;
}
@ Error: Non-exhaustive match: missing variant 'Rectangulo' of enum 'Forma'
```

Si todas las ramas terminan con `return`, el `match` cuenta como que siempre
retorna, como en `area`.

Los nombres ligados en un patrón solo se ven en su rama, y pueden tapar a una
variable de afuera. Si el cuerpo es un bloque, comparte el scope con esos
nombres, como el cuerpo de una función con sus parámetros, así que no se pueden
volver a declarar adentro.

Las reglas sobre el nombre de un enum (con qué choca y qué tapa) están en
[Nombres y alcance en Ricol](scopes_and_names.md). Los nombres de las variantes
no ocupan nombres en el scope.
