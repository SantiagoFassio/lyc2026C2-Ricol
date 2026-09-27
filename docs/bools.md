# Booleanos en Ricol

Los booleanos, en general, son representaciones de expresiones como falsos o verdaderos. Las representaciones de un booleano aparecen al comparar dos expresiones mediante un operador.

Por ejemplo: Comparar la igualdad de dos expresiones resulta en un booleano falso o verdadero segun la igualdad de las partes.

los booleanos tambien se pueden declarar de la siguiente forma: `False;`, `True;`.

## Operadores booleanos

Los siguientes operadores tienen como resultado un booleano, son usadas a modo de comparacion entre expresiones.

### Igualdad, Desigualdades

Estos operadores comparan dos expresiones. Estas expresiones requieren ser del mismo tipo para funcionar. Es decir, que solamente se pueden comparar Strings con Strings y Numeros con Numeros (Floats e Ints son considerados como el mismo tipo). Comparar elementos de distinto tipo arroja un error de tipos.
Como aclaracion, `==` asocia a izquierda, así que `a == b == c` es `(a == b) == c`.


- Igualdad: `==` ==> Compara que dos expresiones sean iguales. Ejemplos:
    - `2 == 2` ==> `True`
    - `1 == 2` ==> `False`
    - `(5 + 7) == 12.0` ==> `True` (La comparacion se resuelve post-aritmetica, los parentesis pueden no estar)
    - `10 / 2 == 5` ==> `True`
    - `"Hola" == "Adios"` ==> `False`
    - `"hola" == "Hola"` ==> `False`
    - `False == True` ==> `False`

- Desigualdad: `!=` ==> Busca que dos expresiones sean distintas, es el efecto contrario a la igualdad

- Mayor / Mayor o Igual a: `>` / `>=` ==> Busca que la expresion de la izquierda sea mayor que la de la derecha / mayor o igual que la de la derecha. Ejemplos:
    - `2 > 1` ==> `True`
    - `2 > 2` ==> `False`
    - `2 >= 1` ==> `True`
    - `2 >= 2` ==> `True`
    - Los Strings poseen ordenamiento de forma alfabetica (de menor a mayor: `"A"` a `"z"`)
        - `"a" > "A"` ==> `True`
        - `"z" > "a"` ==> `True`
    - Los Booleanos no tienen ordenamiento, devuelve un error tratar de comparar el ordenamiento de Booleanos.

- Menor / Menor o Igual a: `<` / `<=` ==> Busca que la expresion de la izquierda sea menor que la de la derecha / menor o igual que la de la derecha. Ejemplos:
    - `1 < 2` ==> `True`
    - `2 < 2` ==> `False`
    - `1 <= 2` ==> `True`
    - `2 <= 2` ==> `True`
    - Los Strings poseen ordenamiento de forma alfabetica (de menor a mayor: `"A"` a `"z"`)
        - `"a" > "A"` ==> `False`
        - `"z" > "a"` ==> `False`
    - Los Booleanos no tienen ordenamiento, devuelve un error tratar de comparar el ordenamiento de Booleanos.

### Nota de importancia

Por prioridad, se resuelven las operaciones de comparacion por orden previo a las comparaciones por igualdad. Ejemplo:

- `True == 1 < 2` ==> `True == (1 < 2)`

### NOT

Expresion que invierte la polaridad del resultado de una expresion booleana, se procesa despues de los operadores de igualdad.
Ejemplos:
- `not True` ==> `False`
- `not False` ==> `True`
- `not 1 > 2` ==> `not (1 > 2)` ==> `not False` ==> `True`
- `not 2 == 2` ==> `not (2 == 2)` ==> `not True` ==> `False`

Not requiere que su parametro sea un booleano.

### OR

Expresion que devuelve `True` si alguno de sus dos parametros (izq o der) es verdadera. Caso contrario devuelve `False`. Se procesa posterior a la operacion `not`.
Ejemplos:
- `True or False` ==> `True`
- `True or True` ==> `True`
- `False or True` ==> `True`
- `False or False` ==> `False`
- `1 == 1 or not 2 < 1` ==> `(1 == 1) or (not (2 < 1))` ==> `True`

### AND

Expresion que devuelve `True` si ambos parametros son verdaderos. Caso contrario devuelve `False`. Se procesa posterior a la operacion `or`.
Ejemplos:
- `True and False` ==> `False`
- `True and True` ==> `True`
- `False and True` ==> `False`
- `False and False` ==> `False`
- `1 == 1 and not 2 < 1` ==> `(1 == 1) and (not (2 < 1))` ==> `True`
- `1 == 2 or 4 != 4 and not 2 < 1` ==> `((1 == 1) or (4 != 4)) and (not (2 < 1))` ==> `False`
