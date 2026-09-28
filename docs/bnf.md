# Backus-Naur Form de Ricol

```plaintext
program                 --> statement* EOF ;

statement               --> expressionStatement | varDeclarationStatement |
                          printStatement | blockStatement | ifStatement |
                          whileStatement | continueStatement | breakStatement |
                          funcDeclarationStatement | returnStatement ;
expressionStatement     --> expression ";" ;
varDeclarationStatement --> "let" IDENTIFIER ":" type "=" expression ";" ;
printStatement          --> "print" expression ";" ;
blockStatement          --> "{" statement* "}" ;
ifStatement             --> "if" "(" expression ")" blockStatement
                          ( "else" ( blockStatement | ifStatement ) )? ;
whileStatement          --> "while" "(" expression ")" blockStatement ;
continueStatement       --> "continue" ";" ;
breakStatement          --> "break" ";" ;
funcDeclarationStatement --> "func" IDENTIFIER "(" parameters? ")" ( "->" type )?
                          blockStatement ;
parameters              --> IDENTIFIER ":" type ( "," IDENTIFIER ":" type )* ;
returnStatement         --> "return" expression? ";" ;

expression              --> assignment ;
assignment              --> IDENTIFIER "=" assignment | logicOr ;
logicOr                 --> logicAnd ( "or" logicAnd )* ;
logicAnd                --> logicNot ( "and" logicNot )* ;
logicNot                --> "not" logicNot | equality ;
equality                --> comparison (( "==" | "!=" ) comparison)* ;
comparison              --> addition (( "<" | "<=" | ">" | ">=" ) addition)* ;
addition                --> term (( "+" | "-" ) term)* ;
term                    --> factor (( "*" | "/" | "//" | "%" ) factor)* ;
factor                  --> ( "-" ) factor | power ;
power                   --> primary ( "**" factor )? ;
primary                 --> INTEGER | FLOAT | STRING | "True" | "False" |
                              IDENTIFIER | call | "(" expression ")" ;
call                    --> IDENTIFIER "(" arguments? ")" ;
arguments               --> expression ( "," expression )* ;

type                    --> "Int" | "Float" | "String" | "Bool" ;
```
