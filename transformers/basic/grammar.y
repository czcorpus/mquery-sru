%{
package basic
%}

%union{
String string
Node node
}


%token <String> TERM NOT AND OR PROX
%type <Node> node

%left AND OR PROX NOT

%%
start: node {yylex.(*basicTransformer).parseResult = &rootNode{$1}} ;

node:
      TERM {$$ = &termNode{$1} }
    | '(' node ')' { $$ = &parenNode{$2}}
    | NOT node { $$ = &unaryNode{Op: "NOT", Child: $2} }
    | node AND node { $$ = &binaryNode{Op: "AND", Left: $1, Right: $3} }
    | node OR node { $$ = &binaryNode{Op: "OR", Left: $1, Right: $3} }
    | node PROX node { $$ = &binaryNode{Op: "PROX", Left: $1, Right: $3} }
    ;
%%