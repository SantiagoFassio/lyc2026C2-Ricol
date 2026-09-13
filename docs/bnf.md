# Backus-Naur Form de Ricol

```plaintext
program     --> statement* EOF ;
statement   --> expression ";" ;
expression  --> term (( "+" | "-" ) term)* ;
term        --> unary (( "*" | "/" | "//" | "%" ) unary)* ;
unary       --> ( "-" ) unary | power ;
power       --> primary ( "**" unary )? ;
primary     --> INTEGER | FLOAT | "(" expression ")" ;
```
