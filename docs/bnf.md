# Backus-Naur Form de Ricol

```plaintext
program             --> statement* EOF ;

statement           --> expressionStatement | printStatement ;
expressionStatement --> expression ";" ;
printStatement      --> "PRINT" expression ";" ;

expression          --> logicOr ;
logicOr             --> logicAnd ( "or" logicAnd )* ;
logicAnd            --> logicNot ( "and" logicNot )* ;
logicNot            --> "not" logicNot | equality ;
equality            --> comparison (( "==" | "!=" ) comparison)* ;
comparison          --> addition (( "<" | "<=" | ">" | ">=" ) addition)* ;
addition            --> term (( "+" | "-" ) term)* ;
term                --> factor (( "*" | "/" | "//" | "%" ) factor)* ;
factor              --> ( "-" ) factor | power ;
power               --> primary ( "**" factor )? ;
primary             --> INTEGER | FLOAT | STRING | "True" | "False" | "(" expression ")" ;
```
