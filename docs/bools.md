# Booleanos en Ricol

Los booleanos, en general, son representaciones de expresiones como falsos o verdaderos. Las representaciones de un booleano aparecen al comparar dos expresiones mediante un operador.

Por ejemplo: Comparar la igualdad de dos expresiones resulta en un booleano falso o verdadero segun la igualdad de las partes.

los booleanos tambien se pueden declarar de la siguiente forma: `False;`, `True;`.

## Operadores booleanos

Los siguientes operadores tienen como resultado un booleano, son usadas a modo de comparacion entre expresiones.

### Igualdad, Mayor, Menor que...

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
