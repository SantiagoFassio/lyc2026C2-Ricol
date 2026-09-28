# Operaciones aritméticas en Ricol

Las operaciones aritméticas trabajan con operandos `Int` y `Float`. Los
tipos numéricos se describen en [Números en Ricol](numbers.md).

## Operadores

| Operador | Operación | Ejemplo | Resultado |
| --- | --- | --- | --- |
| `+` | Suma | `1 + 2` | `3` |
| `-` | Resta | `10 - 2.5` | `7.5` |
| `*` | Multiplicación | `2 * 3` | `6` |
| `/` | División | `10 / 4` | `2.5` |
| `//` | División entera (redondeo hacia abajo) | `10 // 4` | `2` |
| `%` | Resto | `10 % 4` | `2` |
| `**` | Potencia | `2 ** 10` | `1024` |
| `-` (unario) | Opuesto | `-(-3)` | `3` |

## Tipo del resultado

Si los dos operandos son `Int`, el resultado es `Int`, salvo en la
división `/`, que siempre devuelve `Float`. Si algún operando es `Float`, el
`Int` se convierte a `Float` y el resultado es `Float`.

| Operador | `Int` y `Int` | Algún operando `Float` |
| --- | --- | --- |
| `+` `-` `*` | `Int` | `Float` |
| `/` | `Float` | `Float` |
| `//` `%` | `Int` | `Float` |
| `**` | `Int` | `Float` |

El `-` unario conserva el tipo de su operando.

El tipo del resultado depende solo de los tipos de los operandos, nunca de sus
valores. Por eso `10 / 5` es un `Float` aunque la división sea exacta, y
`2 ** -1` es un `Int` (y falla al ejecutarse, ver
[Potencia](#potencia)).

## Precedencia y asociatividad

De mayor a menor precedencia:

| Precedencia | Operadores | Asociatividad |
| --- | --- | --- |
| 1 | `( )` | — |
| 2 | `**` | Derecha |
| 3 | `-` (unario) | Derecha |
| 4 | `*` `/` `//` `%` | Izquierda |
| 5 | `+` `-` | Izquierda |

Algunas consecuencias:

```ricol
10 - 2 * 3;     @ =4, la multiplicación se evalúa primero
(10 - 2) * 3;   @ =24, los paréntesis cambian el orden
1 - 2 - 3;      @ =-4, se evalúa como (1 - 2) - 3
100 / 10 / 2;   @ =5, se evalúa como (100 / 10) / 2
2 ** 3 ** 2;    @ =512, se evalúa como 2 ** (3 ** 2)
-2 ** 2;        @ =-4, se evalúa como -(2 ** 2)
2 * 3 ** 2;     @ =18, se evalúa como 2 * (3 ** 2)
```

El `-` unario puede aparecer como operando derecho de cualquier operador
binario, incluido `**`:

```ricol
2 * -3;         @ =-6
2 ** -1.0;      @ =0.5
```

La gramática formal se encuentra en [Backus-Naur Form de Ricol](bnf.md).

## División entera y resto

`//` y `%` redondean el cociente **hacia abajo** (hacia menos infinito). Como
consecuencia, el resto siempre tiene el mismo signo que el divisor:

| Expresión | `//` | `%` |
| --- | --- | --- |
| `7` y `2` | `7 // 2 = 3` | `7 % 2 = 1` |
| `-7` y `2` | `-7 // 2 = -4` | `-7 % 2 = 1` |
| `7` y `-2` | `7 // -2 = -4` | `7 % -2 = -1` |
| `-7` y `-2` | `-7 // -2 = 3` | `-7 % -2 = -1` |

Este es el mismo comportamiento que en Python, por ejemplo. Esto es distinto en
lenguajes como C, Java o Go, los cuales redondean hacia cero (allí `-7 / 2` es
`-3` y `-7 % 2` es `-1`).

Las dos operaciones siempre cumplen la identidad:

```plaintext
a == (a // b) * b + (a % b)
```

Con operandos `Float` las dos operaciones siguen la misma regla y devuelven
`Float`:

```ricol
7.5 // 2;       @ =3.0
-7.5 // 2;      @ =-4.0
-7.5 % 2;       @ =0.5
10 // 0.5;      @ =20.0
```

## Potencia

Si la base y el exponente son `Int`, el resultado es un `Int` exacto:

```ricol
3 ** 34;        @ =16677181699666569
```

Por eso, un exponente `Int` negativo es un error en tiempo de ejecución.
Para obtener un resultado fraccionario, alguno de los operandos debe ser
`Float`:

```ricol
2 ** -1;        @ Error: Cannot raise an integer to a negative power
2 ** -1.0;      @ =0.5
2.0 ** -1;      @ =0.5
```

## Errores en tiempo de ejecución

Hay errores que dependen del **valor** de los operandos y no de su tipo. Por
eso no se detectan al verificar los tipos, sino recién al ejecutar el programa:

| Error | Ejemplo |
| --- | --- |
| División por cero con `/`, `//` o `%` | `5 % 0` |
| Exponente `Int` negativo en `**` | `2 ** -1` |

El programa se detiene en la sentencia que produce el error:

```plaintext
PRINT 5 % 0;
An error occurred: [line 1, column 9] Cannot divide by zero: 5 % 0
```
