# YASP — Yet Another Lisp

## Control flow
### Switch
```
(switch <expr>
  <case 1 expr> <case 1 body>
  <case 2 expr> <case 2 body>
  
  [<optional default>])
```

## Data types
### Product
```
(defstruct ColoredCircleWithData
  (radius Number)
  (color Color)
  data /* arbitrary type */)

((ColoredCircleWithData radius 5 color RED data 42) radius) // 5

(ColoredCircleWithData 5 RED 42)
```
### Disjoint union
#### Enum
```
(defenum Color RED GREEN BLUE)

(switch (Color RED)
  RED 1
  GREEN 2

  5)

(= (Color RED) (Color GREEN)) // false
```

### Structures

## Code representation
```
(defenum YaspType ID NUMBER STRING LIST)
(defstruct YaspNode
  (type YaspType)
  value)

(eval (YaspNode LIST (list (YaspNode ID "+") (YaspNode NUMBER 2) (YaspNode NUMBER 3)))) // 5
```
