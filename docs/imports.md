# Importaciones en Ricol

el comando `import` permite a un archivo Ricol usar elementos declarados en otro archivo Ricol. De esta forma, un cliente y un servidor pueden compartir los tipos de una misma interfaz (ver [Diseño de Ricol](design.md#importación-de-módulos-e-interfaces)).

```ricol
@ cuenta.ric
struct Movimiento {
    monto: Int;
    saldoFinal: Int;
}

enum ResultadoRetiro {
    Ok(mov: Movimiento);
    SaldoInsuficiente(disponible: Int);
}
```

```ricol
@ main.ric
import "cuenta.ric";

let m: Movimiento = Movimiento { monto: 30, saldoFinal: 70 };
print ResultadoRetiro.Ok(m);   @ ResultadoRetiro.Ok(Movimiento { monto: 30, saldoFinal: 70 })
```

A un archivo importado se lo llama **módulo**. Un módulo es un archivo de Ricol comun y corriente. Este archivo tambien puede importar otros modulos.

## Sintaxis

Un import es la palabra reservada `import`, la ruta del módulo como un string y `;`. Los imports van al principio del archivo, antes de cualquier otra sentencia (los comentarios sí pueden ir antes). Un import en cualquier otro lugar, incluso dentro de un bloque o de una función, es un error:

```ricol
print "hola";
import "cuenta.ric";   @ Error: Imports must be at the beginning of the file
```

## Rutas

Una ruta relativa se resuelve desde el directorio del archivo que tiene el import:

```plaintext
app/
├── main.ric          import "lib/cuenta.ric";   →  app/lib/cuenta.ric
└── lib/
    ├── cuenta.ric    import "tipos.ric";        →  app/lib/tipos.ric
    └── tipos.ric
```

Una ruta absoluta se usa tal cual. Si el archivo no existe, es un error:

```ricol
import "noexiste.ric";   @ Error: Cannot import "noexiste.ric": file not found
```

## Qué se importa

Un módulo puede tener cualquier sentencia, no solo declaraciones. Antes de la verificación de tipos, el intérprete arma un único programa: primero van los módulos, en el orden de los imports y cada uno después de los suyos, y al final el archivo principal. Por eso las sentencias de un módulo se ejecutan una sola vez, antes que las del archivo que lo importa:

```ricol
@ saludo.ric
print "cargando saludo.ric\n";

func saludar(nombre: String) {
    print "hola, " + nombre + "\n";
}
```

```ricol
@ main.ric
import "saludo.ric";

print "empieza main.ric\n";
saludar("Ricol");
@ Imprime:
@ cargando saludo.ric
@ empieza main.ric
@ hola, Ricol
```

Como todo queda en el scope global de un único programa:

- Todos los nombres declarados en el scope global de un módulo son visibles en el archivo que lo importa. Lo que se declara dentro de bloques o funciones del módulo sigue siendo local.
- La visibilidad es transitiva: si `main.ric` importa `a.ric` y `a.ric` importa `b.ric`, en `main.ric` también se ven los nombres de `b.ric`. Además, un módulo ve los nombres de los módulos que se cargaron antes que él, aunque no los importe. Sin embargo, es recomendable importar lo que se usa unicamente, sin problema de repetir importaciones en archivos distintos (ver siguiente seccion).
- Dos declaraciones con el mismo nombre en archivos distintos son un error, como si estuvieran en el mismo archivo. Se informa en la segunda, con su archivo:

```ricol
@ main.ric
import "cuenta.ric";

struct Movimiento {   @ Error: [main.ric, line 3, column 8] Struct 'Movimiento' already declared in this scope
    id: Int;
}
```

Dentro de un bloque o de una función, un nombre local puede tapar a uno importado, como a cualquier nombre global (ver [Nombres y alcance en Ricol](scopes_and_names.md)).

## El mismo módulo varias veces

Cada archivo se carga una sola vez, aunque se importe varias veces, desde varios módulos o por rutas distintas que llegan al mismo archivo (por ejemplo `cuenta.ric`, `./cuenta.ric` y `lib/../cuenta.ric`). Las veces siguientes, el import no hace nada: el archivo no se vuelve a leer y sus declaraciones no se duplican.

Para reconocer el mismo archivo se compara su ruta absoluta. Los enlaces simbólicos no se siguen: dos enlaces al mismo archivo cuentan como archivos distintos.

## Ciclos

Un ciclo de imports es un error, que se detecta antes de ejecutar el programa. El mensaje muestra los archivos que forman el ciclo:

```ricol
@ a.ric
import "b.ric";
```

```ricol
@ b.ric
import "a.ric";   @ Error: [b.ric, line 1, column 8] Import cycle: a.ric -> b.ric -> a.ric
```

Un archivo que se importa a sí mismo también es un ciclo.

## Errores

Como un programa puede estar en varios archivos, todos los errores indican en qué archivo ocurrieron, tanto los de la carga de módulos como los de las demás fases:

```plaintext
[malo.ric, line 3, column 8] Undefined type 'Flot'
[division.ric, line 2, column 14] Cannot divide by zero: 1 // 0
```

La ruta del archivo principal es la que se le pasó al intérprete, y la de cada módulo es la ruta resuelta a partir de ella.

Los errores de la carga de módulos se informan de a uno: la carga se detiene en el primero, como el análisis léxico y el sintáctico.

## Carga de módulos

La carga de módulos es una fase del intérprete que ocurre entre el análisis sintáctico del archivo principal y la verificación de tipos:

1. Recorre los imports del archivo, en orden.
2. Si el módulo ya está cargado, sigue con el próximo import. Si se está cargando (es decir, está en el camino de imports que llevó hasta acá), es un ciclo.
3. Si no, lee el archivo, hace su análisis léxico y sintáctico, y carga sus imports de la misma forma (paso 1).
4. Agrega al programa las sentencias del módulo, sin los imports, y lo marca como cargado.

Al final se agregan las sentencias del archivo principal. La verificación de
tipos y la ejecución reciben ese único programa, sin imports.

Los flags `--scan` y `--parse` muestran solo el archivo principal, con sus imports. `--typecheck` y la ejecución normal incluyen los módulos.
