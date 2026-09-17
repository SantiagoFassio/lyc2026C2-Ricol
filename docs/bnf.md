# Backus-Naur Form de Ricol

```plaintext
program     --> statement* EOF ;
statement   --> expression ";" ;
expression  --> term (( "+" | "-" ) term)* ;
term        --> factor (( "*" | "/" | "//" | "%" ) factor)* ;
factor       --> ( "-" ) factor | power ;
power       --> primary ( "**" factor )? ;
primary     --> INTEGER | FLOAT | "(" expression ")" ;
```
