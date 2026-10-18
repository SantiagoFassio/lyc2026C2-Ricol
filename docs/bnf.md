# Backus-Naur Form de Ricol

```plaintext
program                 --> statement* EOF ;

statement                  --> expressionStatement | varDeclarationStatement |
                               printStatement | blockStatement | ifStatement |
                               whileStatement | continueStatement |
                               breakStatement | funcDeclarationStatement |
                               returnStatement | structDeclarationStatement |
                               enumDeclarationStatement | matchStatement;
expressionStatement        --> expression ";" ;
varDeclarationStatement    --> "let" IDENTIFIER ":" type "=" expression ";" ;
printStatement             --> "print" expression ";" ;
blockStatement             --> "{" statement* "}" ;
ifStatement                --> "if" "(" expression ")" blockStatement
                               ( "else" ( blockStatement | ifStatement ) )? ;
whileStatement             --> "while" "(" expression ")" blockStatement ;
continueStatement          --> "continue" ";" ;
breakStatement             --> "break" ";" ;
funcDeclarationStatement   --> "func" IDENTIFIER "(" parameters? ")" ( "->" type )?
                               blockStatement ;
parameters                 --> IDENTIFIER ":" type ( "," IDENTIFIER ":" type )* ;
returnStatement            --> "return" expression? ";" ;
structDeclarationStatement --> "struct" IDENTIFIER "{" fieldDeclaration* "}" ;
fieldDeclaration           --> IDENTIFIER ":" type ";" ;
enumDeclarationStatement   --> "enum" IDENTIFIER "{" variantDeclaration* "}" ;
variantDeclaration         --> IDENTIFIER ( "(" variantFields ")" )? ";" ;
variantFields              --> IDENTIFIER ":" type ( "," IDENTIFIER ":" type )* ;
matchStatement             --> "match" "(" expression ")" "{" matchArm* "}" ;
matchArm                   --> pattern "=>" statement ;
pattern                    --> "_" | IDENTIFIER ( "(" binding ( "," binding )* ")" )? ;
binding                    --> IDENTIFIER | "_" ;


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
power                   --> access ( "**" factor )? ;
access                  --> primary ( "." IDENTIFIER ( "(" arguments? ")" )? )* ;
primary                 --> INTEGER | FLOAT | STRING | "True" | "False" |
                              IDENTIFIER | call | structLiteral | "(" expression ")" ;
call                    --> IDENTIFIER "(" arguments? ")" ;
arguments               --> expression ( "," expression )* ;
structLiteral           --> IDENTIFIER "{" fieldInitializers? "}" ;
fieldInitializers       --> IDENTIFIER ":" expression ( "," IDENTIFIER ":" expression )* ;

type                    --> "Int" | "Float" | "String" | "Bool" | IDENTIFIER ;
```
