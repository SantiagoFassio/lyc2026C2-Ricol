# Backus-Naur Form de Ricol

```plaintext
program             --> statement* EOF ;

statement           --> expressionStatement | printStatement | blockStatement |
                          ifStatement | whileStatement | continueStatement |
                          breakStatement;
expressionStatement --> expression ";" ;
printStatement      --> "PRINT" expression ";" ;
blockStatement      --> "{" statement* "}" ;
ifStatement         --> "if" "(" expression ")" blockStatement
                          ( "else" ( blockStatement | ifStatement ) )? ;
whileStatement      --> "while" "(" expression ")" blockStatement ;
continueStatement   --> "continue" ";" ;
breakStatement      --> "break" ";" ;

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
