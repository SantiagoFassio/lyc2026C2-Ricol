# Números en Ricol

Ricol tiene dos tipos numéricos: `Int` y `Float`. Las operaciones que se
pueden hacer con ellos se describen en
[Operaciones aritméticas en Ricol](arithmetic.md).

## Int

Un `Int` es un número entero de 64 bits con signo. Puede representar los
valores desde `-9223372036854775808` hasta `9223372036854775807`.

### Literales enteros

Un literal entero es una secuencia de dígitos decimales: `0`, `3`, `2026`.

- No existen literales negativos: `-3` es el operador `-` aplicado al literal
  `3`.
- Un literal que no entra en 64 bits es un error:

  ```plaintext
  9223372036854775808;
  An error occurred: [line 1, column 1] Invalid integer: 9223372036854775808
  ```

### Overflow

Si el resultado de una operación entre enteros no entra en 64 bits, sucede un
overflow de forma implícita:

```ricol
print 9223372036854775807 + 1;  @ Imprime -9223372036854775808
```

## Float

Un `Float` es un número de punto flotante de 64 bits (doble precisión) según el
estándar IEEE 754.

### Literales flotantes

Un literal flotante es una secuencia de dígitos con un punto decimal: `41.3`,
`0.5`, `2.0`.

- El punto puede ir al final (`3.` es el Float `3.0`), pero no al principio:
  `.5` es un error, hay que escribir `0.5`.
- No hay notación científica: `1e3` es un error.
- Un número con más de un punto (`3.23.4`) es un error.

### Diferencias con IEEE 754

Ricol sigue el estándar IEEE 754, con una excepción: **dividir por cero es un
error**, también cuando los operandos son `Float`. El estándar indica que
`1.0 / 0.0` resulte en `+Inf`, pero en Ricol el programa termina con un error.
Lo mismo ocurre con `//` y `%`:

```plaintext
print 1.0 / 0.0;
An error occurred: [line 1, column 11] Cannot divide by zero: 1 / 0
```

Como en cualquier implementación de IEEE 754, algunos resultados no son exactos:

```ricol
print 0.1 + 0.2;     @ Imprime 0.30000000000000004
```

## Conversión entre Int y Float

La única conversión implícita de Ricol es de `Int` a `Float`. Ocurre cuando
una operación aritmética combina un `Int` con un `Float`: el `Int` se
convierte a `Float` y el resultado es un `Float`.

```ricol
1 + 2.3;   @ Float 3.3
```

Nunca se convierte un `Float` a `Int` de forma implícita.
