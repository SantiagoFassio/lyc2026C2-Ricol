# Nombres y alcance en Ricol

En Ricol, las variables, los parámetros y las funciones comparten los nombres: no hay un espacio de nombres para las variables y otro para las funciones. Un mismo nombre puede referirse a una variable en un lugar del programa y a una función en otro, pero nunca a las dos cosas en el mismo scope.

Todas las reglas de este documento se verifican antes de ejecutar el programa, en la fase de verificación de tipos.

## Scopes

Un scope es la región del programa en la que un nombre declarado es visible. Cuando se declara algo en un scope, otras secciones de código dentro del scope pueden usarlo mediante el nombre dado (incluyendo scopes dentro del scope mayor, ver scopes anidados).

Ricol tiene alcance léxico: los scopes se deducen de la estructura del código, no del orden en que se ejecuta. Hay tres tipos de scope:

- **Global**: el programa completo.
- **Bloque**: cada `{}`, tanto los bloques sueltos como las ramas de un `if` y el cuerpo de un `while`.
- **Función**: los parámetros y el cuerpo de una función forman **un único** scope. Por eso, una variable local no puede llamarse igual que un parámetro.

### Scopes anidados

Los scopes se anidan. Al averiguar a que se refiere un nombre (de una función que se llama o de una variable que se usa), se busca desde el scope actual hacia afuera, y se usa la **primera** declaración que se encuentra, sin importar si es una variable o una función.

Esto causa que los nombres declarados dentro de un scope toman prioridad por encima de los scopes "mayores" si estos nombres son llamados dentro del mismo scope o uno por debajo de la cadena de anidación (ver "shadowing").

## Mismo scope: error

Declarar dos veces el mismo nombre en un scope es un error, aunque una
declaración sea una variable y la otra una función:

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

## Errores, usar nombres de forma incorrecta

Usar un nombre para algo que no es genera los siguientes errores:

| Uso | Error |
| :--- | :--- |
| Llamar a una variable o a un parámetro | `'x' is not a function` |
| Usar una función como valor | `Cannot use function 'f' as a value` |
| Asignarle un valor a una función | `Cannot assign to function 'f'` |

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

El typechecker guarda, por cada scope, una única tabla con todos los nombres declarados en él ([scope_stack.go](../src/typechecker/scope_stack.go)). Cada entrada es una variable, con su tipo, o una función, con su firma. Que la tabla sea única es lo que hace que una variable y una función no puedan compartir nombre en un mismo scope.

El intérprete guarda las variables y las funciones en dos tablas separadas, pero no puede haber un conflicto de nombres por un paso previo. El typechecker ya garantizó que cada nombre tenga una sola declaración por scope y que se use según su tipo.

## Pruebas

- [integration/8-names.ric](../integration/8-names.ric) ejecuta los casos válidos de este documento.
- Los casos de error están cubiertos por los tests del chequeador de tipos en [typechecker_error_test.go](../src/typechecker/typechecker_error_test.go): `TestFunctionRedeclaration`, `TestCallErrors` y `TestFunctionDeclarationOrder`.
