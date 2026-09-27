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
