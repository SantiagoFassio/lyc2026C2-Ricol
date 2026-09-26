# Backus-Naur Form de Ricol

```plaintext
program     --> statement* EOF ;
statement   --> expression ";" ;
expression  --> equality ;
equality    --> comparison (( "==" | "!=" ) comparison)* ;
comparison  --> addition (( "<" | "<=" | ">" | ">=" ) addition)* ;
addition    --> term (( "+" | "-" ) term)* ;
term        --> factor (( "*" | "/" | "//" | "%" ) factor)* ;
factor       --> ( "-" ) factor | power ;
power       --> primary ( "**" factor )? ;
primary     --> INTEGER | FLOAT | STRING | "True" | "False" | "(" expression ")" ;
```
