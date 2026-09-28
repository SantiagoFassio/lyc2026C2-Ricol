# Resultados de los benchmarks

Registro de una corrida de `make bench`.

## Lenguajes

| Lenguaje | Implementación | Versión y compilación |
| :--- | :--- | :--- |
| C | GCC (compilador a código nativo) | gcc (Debian 14.2.0-19) 14.2.0, target `x86_64-linux-gnu`, flags `-O2`, glibc 2.41 |
| Go | gc (compilador oficial a código nativo) | go1.27.1 linux/amd64, `GOAMD64=v1`, `go build` sin flags |
| Python | Intérprete de bytecode | CPython 3.13.5, compilado con GCC 14.2.0; JIT: no; GIL: sí |
| Ricol | Intérprete tree-walking escrito en Go | compilado con el mismo Go que la fila anterior |

## Entorno

- CPU: Intel(R) Core(TM) Ultra 7 155H, núcleos asignados: 1
- Kernel: Linux 7.0.0-34-generic, contenedor Docker, Debian GNU/Linux 13 (trixie)
- hyperfine 1.19.0, con 2 corridas de calentamiento y 10 corridas medidas por programa, sin shell intermedio (`--shell=none`)

## calls

| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
| :--- | ---: | ---: | ---: | ---: |
| `C` | 3.4 ± 0.0 | 3.3 | 3.4 | 1.00 |
| `Go` | 3.4 ± 0.1 | 3.3 | 3.5 | 1.01 ± 0.02 |
| `Python` | 134.7 ± 3.6 | 130.3 | 140.8 | 39.77 ± 1.16 |
| `Ricol` | 838.0 ± 5.5 | 832.2 | 847.1 | 247.44 ± 3.28 |

## deep_rec

| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
| :--- | ---: | ---: | ---: | ---: |
| `C` | 1.0 ± 0.1 | 1.0 | 1.2 | 1.00 |
| `Go` | 12.0 ± 0.4 | 11.7 | 13.0 | 11.85 ± 0.91 |
| `Python` | 233.6 ± 1.8 | 231.0 | 237.6 | 231.52 ± 16.00 |
| `Ricol` | 2426.9 ± 18.2 | 2397.2 | 2455.9 | 2404.99 ± 166.13 |

## empty

| Command | Mean [µs] | Min [µs] | Max [µs] | Relative |
| :--- | ---: | ---: | ---: | ---: |
| `C` | 264.2 ± 8.3 | 253.3 | 283.9 | 1.00 |
| `Go` | 497.5 ± 9.5 | 481.8 | 512.9 | 1.88 ± 0.07 |
| `Python` | 7549.3 ± 207.0 | 7307.5 | 7879.3 | 28.58 ± 1.19 |
| `Ricol` | 586.3 ± 45.7 | 549.1 | 670.4 | 2.22 ± 0.19 |

## fib

| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
| :--- | ---: | ---: | ---: | ---: |
| `C` | 1.4 ± 0.0 | 1.4 | 1.5 | 1.00 |
| `Go` | 4.1 ± 0.2 | 3.9 | 4.4 | 2.82 ± 0.12 |
| `Python` | 88.1 ± 1.4 | 86.6 | 91.1 | 61.09 ± 1.37 |
| `Ricol` | 1290.4 ± 10.1 | 1275.9 | 1308.7 | 894.54 ± 15.85 |

## loop

| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
| :--- | ---: | ---: | ---: | ---: |
| `C` | 18.0 ± 0.6 | 17.2 | 19.0 | 1.08 ± 0.05 |
| `Go` | 16.7 ± 0.4 | 16.1 | 17.3 | 1.00 |
| `Python` | 509.5 ± 2.6 | 506.0 | 514.8 | 30.58 ± 0.77 |
| `Ricol` | 1907.9 ± 8.0 | 1894.5 | 1923.2 | 114.51 ± 2.85 |

## primes

| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
| :--- | ---: | ---: | ---: | ---: |
| `C` | 18.9 ± 0.5 | 18.1 | 19.8 | 1.08 ± 0.04 |
| `Go` | 17.5 ± 0.4 | 17.0 | 18.0 | 1.00 |
| `Python` | 397.1 ± 3.6 | 392.6 | 403.3 | 22.71 ± 0.51 |
| `Ricol` | 2703.4 ± 9.7 | 2688.9 | 2718.6 | 154.63 ± 3.24 |

## Conclusiones

### Ricol frente a Python

Ricol es entre 3.7 y 14.6 veces más lento que CPython. La distancia depende de
cuántas llamadas a funciones hace el programa:

| Benchmark | Ricol / Python | Qué predomina |
| :--- | ---: | :--- |
| `loop` | 3.7× | Aritmética y asignaciones dentro de un `while` |
| `calls` | 6.2× | Una llamada corta por iteración |
| `primes` | 6.8× | Loops anidados, con una llamada por número |
| `deep_rec` | 10.4× | Llamadas con pila muy profunda |
| `fib` | 14.6× | Solo llamadas |

Tomando solo el tiempo de ejecución, sin el arranque, una iteración de `loop`
cuesta unos 380 ns en Ricol contra 100 ns en Python. Una llamada de `fib`
cuesta 480 ns contra 33 ns: las llamadas a funciones son el principal punto
débil de Ricol.

El arranque de Ricol, en cambio, es barato. `empty` tarda 0.59 ms, casi lo
mismo que un binario de Go vacío (0.50 ms) y unas 13 veces menos que CPython
(7.5 ms). El scanner, el parser y el typechecker no son un problema.

### Dónde se va el tiempo en Ricol

Los perfiles muestran que el costo no está en la aritmética, sino en las
estructuras que el intérprete crea y recorre. Entre ellas se encuentran los
environments nuevos que se crean en cada llamada a función, los números que se
utilizan mediante una interfaz, el `return` implementado como `error` y el
`scopeStack`.

### Limitaciones

- Todas las mediciones son de una sola máquina, usando un solo núcleo dentro
  de Docker. Los tiempos absolutos cambian entre máquinas; las proporciones
  entre lenguajes son más estables.
- CPython 3.13 corre sin JIT. Con PyPy u otro intérprete con JIT, la
  comparación contra Python cambiaría mucho.
- `deep_rec` usa profundidades menores a 10000 porque ese es el límite de
  llamadas anidadas de Ricol (`maxCallDepth`). No mide recursiones más
  profundas en ningún lenguaje.
