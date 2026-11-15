# Optimizaciones de Ricol

Antes de ejecutar un programa, Ricol lo pasa por una fase llamada
**optimizador**, que reescribe el AST sin cambiar lo que el programa hace. El
optimizador aplica tres transformaciones: constant folding, dead code
elimination y eliminación de agrupaciones.

El efecto de cada una está medido en
[Resultados de las optimizaciones](../benchmarks/results/results_summary.md).

## Índice

- [El optimizador](#el-optimizador)
- [Constant folding](#constant-folding)
- [Dead code elimination](#dead-code-elimination)
- [Eliminación de agrupaciones](#eliminación-de-agrupaciones)
- [Lo que el optimizador no hace](#lo-que-el-optimizador-no-hace)

## El optimizador

El optimizador es una fase que corre después de la verificación de tipos y
antes de la ejecución:

```plaintext
scanner → parser → loader → typechecker → optimizador → intérprete
```

Corre después del typechecker por dos motivos: solo optimiza programas que ya
se sabe que son válidos, y puede confiar en los tipos (por ejemplo, que la
condición de un `if` es un `Bool`). Optimiza cada módulo por separado,
recorriendo todas sus sentencias y expresiones, incluidos los cuerpos de
funciones, los brazos de `match`, los servers y las sesiones.

El flag `--optimize` muestra el archivo principal tal como queda después de
esta fase. Por ejemplo, este programa:

```ricol
let horas: Int = 3;
while (horas > 0) {
    if (1 > 2) {
        print "nunca";
    }
    print horas * (60 * 60);
    print "\n";
    horas = horas - 1;
}
```

queda así:

```plaintext
$ go run . --optimize programa.ric
let horas : Int = INTEGER<3>;
while ((horas GREATER<>> INTEGER<0>)) {
  print (horas STAR<*> INTEGER<3600>);
  print STRING<"\n">;
  (horas = (horas MINUS<-> INTEGER<1>));
}
```

`60 * 60` se calculó una sola vez (constant folding), el `if (1 > 2)` se
eliminó (dead code elimination) y los paréntesis ya no aparecen como nodos
(eliminación de agrupaciones). Los paréntesis que se ven en la salida son los
que `--optimize` y `--parse` escriben alrededor de cada operación, no nodos del
AST.

Las tres transformaciones se aplican en una sola pasada, de abajo hacia arriba:
primero se optimizan los hijos de un nodo y después el nodo. Así una
transformación habilita a otra: el folding convierte `1 > 2` en `False`, y
entonces la dead code elimination puede eliminar el `if`.

## Constant folding

Calcula antes de ejecutar las expresiones cuyos operandos son todos literales,
y las reemplaza por su resultado.

| Expresión | Resultado |
| :--- | :--- |
| Binaria con los dos operandos literales (`60 * 60`, `"a" + "b"`, `1 < 2`) | El literal con el resultado |
| Unaria con operando literal (`-5`, `not True`) | El literal |
| `False and x`, `True or x` | El literal izquierdo, porque `x` nunca se evalúa |
| `True and x`, `False or x` | `x` |

Para que el resultado sea idéntico al de la ejecución, el optimizador no
reimplementa los operadores: le pide al intérprete que evalúe la expresión. Así
se respetan la división entera hacia abajo, el desborde de los enteros y las
reglas entre `Int` y `Float`.

Una operación que falla no se pliega. `1 / 0` o `2 ** -1` quedan como están,
para que el error siga ocurriendo al ejecutar el programa, con la misma
posición. Por ejemplo, `print 10 / (5 - 5);` queda como:

```plaintext
print (INTEGER<10> SLASH</> INTEGER<0>);
```

## Dead code elimination

Elimina el código que nunca se va a ejecutar:

| Sentencia | Resultado |
| :--- | :--- |
| `if (False) A` | Se elimina |
| `if (False) A else B` | `B` |
| `if (True) A else B` | `A` |
| `while (False) A` | Se elimina |
| Sentencias después de `return`, `break` o `continue` en el mismo bloque | Se eliminan |

Las ramas que quedan se conservan como bloques, así las variables declaradas
adentro siguen en su propio scope. Si el cuerpo de un `while`, un brazo de
`match` o la rama de un `if` se queda sin sentencias, se reemplaza por un
bloque vacío.

```ricol
func signo(n: Int) -> Int {
    if (n < 0) {
        return -1;
        print "inalcanzable";
    }
    return 1;
}
```

```plaintext
func signo(n : Int) -> Int {
  if ((n LESS<<> INTEGER<0>)) {
    return INTEGER<-1>;
  }
  return INTEGER<1>;
}
```

Como Ricol no tiene constantes y todas las variables se pueden reasignar, la
dead code elimination solo encuentra ramas muertas cuando la condición está
escrita con literales.

## Eliminación de agrupaciones

Cada par de paréntesis crea un nodo `GroupingExpression` en el AST. Los
paréntesis deciden la forma del árbol, pero una vez armado el nodo no aporta
nada, y evaluarlo le cuesta al intérprete una pasada más. El optimizador
reemplaza cada agrupación por la expresión que contiene.

Se hace en el optimizador y no en el parser para que `--parse` siga mostrando
los paréntesis tal como se escribieron.

## Lo que el optimizador no hace

- **Propagación de constantes**: no reemplaza una variable que nunca se
  reasigna por su valor. Por lo tanto, en casos como `let debug: Bool = False; if (debug) {...}`
  no se elimina la variable.
- **Reordenar operaciones**: `x + 1 + 2` se lee como `(x + 1) + 2` y no se
  pliega, porque reordenar no es seguro en general con floats. Hay que
  escribir `x + (1 + 2)`.
- **Aplanar bloques**: el bloque que queda de un `if (True)` sigue creando su
  environment al ejecutarse.
- **Eliminar declaraciones sin usar**: una función, un struct o un enum que
  nadie usa se mantiene. No cambia el tiempo de ejecución.
