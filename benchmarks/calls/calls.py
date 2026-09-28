# Many calls to a short function: measures the cost of each call
def step(acc, i):
    return (acc + i * 7) % 1000003


acc = 0
i = 0
while i < 1000000:
    acc = step(acc, i)
    i = i + 1
print(acc)
