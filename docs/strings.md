# Strings en Ricol

Un string en Ricol es identificado por encontrarse entre dos comillas dobles (Ejemplo: `"Por favor denme un 10"`).

## Scanner

### Reconocimiento

- Un string empieza con `"` y termina en la siguiente `"` que no esté escapada. Esto solo genera un Token.
- El lexema es el texto tal como aparece en el código fuente. No se identifican elementos como cambios de linea dentro del scanner.
- El scanner no calcula el valor del string. Quitar las comillas y traducir los escapes le corresponde al parser.

### Contenido permitido

- Strings vacíos `""`.
- Otros identificadores de codigo, operadores, etc. `@` no inicia un comentario. Operadores como `+`, `=`, y finales de linea (`;`) son tomados como parte del string.
- Unicode

### Secuencias de escape

| Secuencia | Significado |
|---|---|
| `\"` | Comilla doble |
| `\\` | Barra invertida |
| `\n` | Salto de línea |
| `\t` | Tabulación |

El scanner si esta encargado de detectar que los caracteres posteriores a una barra invertida sean correspondientes con alguna de las secuencias de escape.

### Errores léxicos

| Error | Situacion |
|---|---|
| String sin cerrar | Se llega al final de la línea o del archivo sin encontrar la doble comilla de cierre |
| Secuencia de escape inválida | Después de una `\` viene un carácter que no está en la tabla de escapes |

Los strings no pueden ocupar más de una línea.

El error por secuencia de escape invalida toma prioridad por encima del string sin cerrar por el sencillo motivo de que la secuencia invalida se encuentra antes.

### Operaciones con operadores matematicos

El unico operador funcional con strings es el simbolo `+`, el cual funciona concatenando dos strings.

Es posible concatenar strings de la forma `"Buenos " + "dias"`. Usar otro tipo de dato en la izquierda o derecha de la expresion resultara en error.

### Ejemplos

Ejemplos se encuentran bajo la carpeta examples/
- strings_scanner.ric
- strings_scanner_error.ric
- strings_parser.ric

