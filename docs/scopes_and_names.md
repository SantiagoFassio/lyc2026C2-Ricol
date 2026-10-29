# Nombres y alcance en Ricol

En Ricol, las variables, los parámetros, las funciones, los structs y los enums comparten los nombres: no hay un espacio de nombres para cada tipo de declaración. Un mismo nombre puede referirse a una variable en un lugar del programa y a una función, a un struct o a un enum en otro, pero nunca a dos de esas cosas en el mismo scope. Los nombres de las variantes de un enum no ocupan nombres en el scope, porque viven dentro del enum.

Todas las reglas de este documento se verifican antes de ejecutar el programa, en la fase de verificación de tipos.

## Scopes

Un scope es la región del programa en la que un nombre declarado es visible. Cuando se declara algo en un scope, otras secciones de código dentro del scope pueden usarlo mediante el nombre dado (incluyendo scopes dentro del scope mayor, ver scopes anidados).

Ricol tiene alcance léxico: los scopes se deducen de la estructura del código, no del orden en que se ejecuta. Hay tres tipos de scope:

- **Global**: el programa completo.
- **Bloque**: cada `{}`, tanto los bloques sueltos como las ramas de un `if` y el cuerpo de un `while`.
- **Función**: los parámetros y el cuerpo de una función forman **un único** scope. Por eso, una variable local no puede llamarse igual que un parámetro.

### Scopes anidados

Los scopes se anidan. Al averiguar a que se refiere un nombre (de una función que se llama o de una variable que se usa), se busca desde el scope actual hacia afuera, y se usa la **primera** declaración que se encuentra, sin importar si es una variable, una función, un struct o un enum.

Esto causa que los nombres declarados dentro de un scope toman prioridad por encima de los scopes "mayores" si estos nombres son llamados dentro del mismo scope o uno por debajo de la cadena de anidación (ver "shadowing").

## Mismo scope: error

Declarar dos veces el mismo nombre en un scope es un error, aunque sean
declaraciones de distinto tipo (por ejemplo, una variable y una función):

```ricol
func f() -> Int {
    return 1;
}
let f: Int = 2;     @ Error: Variable 'f' already declared in this scope
```

```ricol
let f: Int = 2;
func f() -> Int {   @ Error: Function 'f' already declared in this scope
    return 1;
}
```

```ricol
let Punto: Int = 2;
struct Punto {      @ Error: Struct 'Punto' already declared in this scope
    x: Int;
}
```

```ricol
struct Color {
    nombre: String;
}
enum Color {        @ Error: Enum 'Color' already declared in this scope
    Rojo;
}
```

## Scopes distintos: shadowing

Una declaración en un scope interno tapa a la del scope externo con el mismo nombre, aunque sean de distinto tipo. Al salir del scope interno, el nombre vuelve a referirse a la declaración externa. Esto es denomindo _shadowing_.

```ricol
func valor() -> Int {
    return 1;
}
{
    let valor: Int = 5;
    print valor;        @ 5
    valor();            @ Error: 'valor' is not a function
}
print valor();          @ 1
```

```ricol
let numero: Int = 1;
{
    func numero() -> Int {
        return 2;
    }
    print numero();     @ 2
    print numero;       @ Error: Cannot use function 'numero' as a value
}
print numero;           @ 1
```

Lo mismo pasa con los parámetros y las variables locales de una función, que pueden llamarse igual que la función o que otra función de afuera. Hacer esto (declarar un parametro con el mismo nombre que la función) causa que no se puedan hacer llamadas recursivas sobre esa función:

```ricol
func por10(por10: Int) -> Int {
    return por10 * 10;  @ por10 es el parámetro
}
print por10(3);         @ 30

func cuenta(cuenta: Int) -> Int {
    return cuenta(cuenta - 1);  @ Error: 'cuenta' is not a function
}
```

Con los structs pasa lo mismo: una variable o una función interna puede tapar a un struct, y mientras esté tapado su nombre no se puede usar como tipo ni para construir valores. Al revés, un struct interno también puede tapar a una variable o a una función externa.

```ricol
struct Punto {
    x: Int;
}
{
    let Punto: Int = 5;
    print Punto;                @ 5
    func f(p: Punto) {}         @ Error: 'Punto' is not a type
    print Punto { x: 1 };       @ Error: 'Punto' is not a struct
}
print Punto { x: 1 };           @ Punto { x: 1 }
```

Si un struct interno tapa a otro struct con el mismo nombre, son dos tipos distintos: un valor de uno no se puede usar donde se espera el otro, aunque los mensajes de error muestren el mismo nombre (por ejemplo, `Cannot assign Punto to variable of type Punto`).

Con los enums pasa lo mismo. Mientras un enum está tapado, su nombre no se puede usar como tipo ni para construir variantes: `Color.Rojo` pasa a ser un acceso al campo `Rojo` de lo que lo tapa.

```ricol
enum Color {
    Rojo;
}
{
    let Color: Int = 1;
    print Color.Rojo;           @ Error: Cannot access field 'Rojo' of type Int
}
print Color.Rojo;               @ Color.Rojo
```

## Errores, usar nombres de forma incorrecta

Usar un nombre para algo que no es genera los siguientes errores:

| Uso | Error |
| :--- | :--- |
| Llamar a una variable o a un parámetro | `'x' is not a function` |
| Usar una función como valor | `Cannot use function 'f' as a value` |
| Asignarle un valor a una función | `Cannot assign to function 'f'` |
| Usar un struct como valor | `Cannot use struct 'P' as a value` |
| Asignarle un valor a un struct | `Cannot assign to struct 'P'` |
| Llamar a un struct | `'P' is not a function` |
| Usar un enum como valor | `Cannot use enum 'E' as a value` |
| Asignarle un valor a un enum | `Cannot assign to enum 'E'` |
| Llamar a un enum | `'E' is not a function` |
| Construir un valor de struct con el nombre de un enum | `'E' is not a struct` |
| Usar una variable o una función como tipo | `'x' is not a type` |
| Construir un valor con el nombre de una variable o una función | `'x' is not a struct` |

## Desde dónde es visible un nombre

Una variable es visible desde la sentencia siguiente a su declaración. Su valor inicial se calcula antes de declararla, así que puede usar la declaración externa con el mismo nombre:

```ricol
func base() -> Int {
    return 1;
}
{
    let base: Int = base() + 10;    @ base() es la función de afuera
    print base;                     @ 11
}
```

Una función es visible desde su propia declaración para permitir recursividad. Además, las declaraciones `func` **seguidas** forman un grupo, y todas las funciones del grupo se ven entre sí, permitiendo recursión mutua. Si entre dos funciones hay otra sentencia, forman dos grupos distintos y la primera no ve a la segunda.

Un struct, igual que una variable, es visible desde la sentencia siguiente a su declaración. Los tipos de sus campos se resuelven antes de declararlo, así que un struct no puede contenerse a sí mismo, pero un campo sí puede usar un struct externo con el mismo nombre:

```ricol
struct Nodo {
    siguiente: Nodo;        @ Error: Undefined type 'Nodo'
}
```

```ricol
struct Version {
    numero: Int;
}
{
    struct Version {
        anterior: Version;  @ el struct Version de afuera
        numero: Int;
    }
}
```

Los structs no forman grupos como las funciones, así que una función declarada antes que un struct no puede usarlo en su firma:

```ricol
func origen(p: Punto) {}    @ Error: Undefined type 'Punto'
struct Punto {
    x: Int;
}
```

Un enum, a diferencia de un struct, es visible **desde su propia declaración**: se declara antes de resolver los tipos de los campos de sus variantes. Eso es lo que permite que un enum se use a sí mismo (ver [Enums en Ricol](enums.md#declaración-de-un-enum)). Como consecuencia, si un enum interno tiene el mismo nombre que un tipo externo, sus campos usan el enum interno y no el externo:

```ricol
enum Lista {
    Vacia;
    Nodo(valor: Int, resto: Lista);     @ resto es de la propia Lista
}
```

Fuera de eso, los enums se comportan como los structs: no forman grupos, así que una función o un struct declarados antes que un enum no pueden usarlo.

## Resolución de nombres de funciones

Al declarar una función, se resuelven los nombres al momento de la declaración, por lo que si se tapa el nombre usado por la función con otra declaración y luego se llama a la función esta usara los nombres declarados originalmente, no los tapados.

```ricol
func origen() -> String {
    return "función";
}
{
    func leer() -> String {
        return origen();
    }
    let origen: String = "variable";
    print leer();       @ leer() usa la función origen(), por lo que imprime "función".
    print origen;       @ imprime "variable".
}
```

## Implementación

El typechecker guarda, por cada scope, una única tabla con todos los nombres declarados en él ([scope_stack.go](../src/typechecker/scope_stack.go)). Cada entrada es una variable, con su tipo; una función, con su firma; un struct, con su tipo (`*types.StructType`); o un enum, con su tipo (`*types.EnumType`). Que la tabla sea única es lo que hace que dos declaraciones no puedan compartir nombre en un mismo scope.

El intérprete guarda las variables, las funciones, los structs y los enums en cuatro tablas separadas, pero no puede haber un conflicto de nombres por un paso previo. El typechecker ya garantizó que cada nombre tenga una sola declaración por scope y que se use según su tipo. Eso es lo que usa el intérprete para distinguir la construcción de una variante (`Color.Rojo`) de un acceso a campo (`p.x`): mira si el nombre de la izquierda está en la tabla de enums del scope que le corresponde.

## Pruebas

- [integration/8-names.ric](../integration/8-names.ric) ejecuta los casos válidos de este documento, [integration/9-structs.ric](../integration/9-structs.ric) los de los structs y [integration/10-enums.ric](../integration/10-enums.ric) los de los enums.
- Los casos de error están cubiertos por los tests del chequeador de tipos en [typechecker_error_test.go](../src/typechecker/typechecker_error_test.go): `TestFunctionRedeclaration`, `TestCallErrors` y `TestFunctionDeclarationOrder`. Los de los structs, en `TestStructDeclarationErrors`, `TestNamedTypeErrors` y `TestStructNameErrors`. Los de los enums, en `TestEnumDeclarationErrors` y `TestEnumNameErrors`.
