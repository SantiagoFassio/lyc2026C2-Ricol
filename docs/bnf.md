# Backus-Naur Form de Ricol

```plaintext
program             --> statement* EOF ;

statement           --> expressionStatement | printStatement ;
expressionStatement --> expression ";"
printStatement      --> "PRINT" expression ";" ;

expression          --> term (( "+" | "-" ) term)* ;
term                --> factor (( "*" | "/" | "//" | "%" ) factor)* ;
factor              --> ( "-" ) factor | power ;
power               --> primary ( "**" factor )? ;
primary             --> INTEGER | FLOAT | STRING | "(" expression ")" ;
```
