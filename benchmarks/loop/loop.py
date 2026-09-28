# Loop with many iterations and arithmetic the compiler cannot evaluate at compile time
i = 0
acc = 0
while i < 5000000:
    acc = (acc * 31 + i) % 1000003
    i = i + 1
print(acc)
