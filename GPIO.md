# GPIO

| Pin Number | Direction (Input/Output) | Function/Usage   |
| ---------- | ------------------------ | ---------------- |
| 0          | Input                    | Reset Button     |
| 1          | Input                    | WPS Button       |
| 2          | Output                   | Blue Mid LED     |
| 3          | Output                   | Green Bottom LED |
| 13         | Output                   | Green Mid LED    |
| 15         | Output                   | Red Mid LED      |
| 16         | Input                    | Check Button     |

## LEDs

### white

```bash
echo high > /sys/class/gpio/gpio2/direction
echo high > /sys/class/gpio/gpio13/direction
echo high > /sys/class/gpio/gpio15/direction
```

### blue

```bash
echo high > /sys/class/gpio/gpio2/direction
echo low > /sys/class/gpio/gpio13/direction
echo low > /sys/class/gpio/gpio15/direction
```

### green

```bash
echo low > /sys/class/gpio/gpio2/direction
echo high > /sys/class/gpio/gpio13/direction
echo low > /sys/class/gpio/gpio15/direction
```

### red

```bash
echo low > /sys/class/gpio/gpio2/direction
echo low > /sys/class/gpio/gpio13/direction
echo high > /sys/class/gpio/gpio15/direction
```

### off

```bash
echo low > /sys/class/gpio/gpio2/direction
echo low > /sys/class/gpio/gpio13/direction
echo low > /sys/class/gpio/gpio15/direction
```

### yellow

```bash
echo low > /sys/class/gpio/gpio2/direction
echo high > /sys/class/gpio/gpio13/direction
echo high > /sys/class/gpio/gpio15/direction
```

### purple

```bash
echo high > /sys/class/gpio/gpio2/direction
echo low > /sys/class/gpio/gpio13/direction
echo high > /sys/class/gpio/gpio15/direction
```

### blue bright

```bash
echo high > /sys/class/gpio/gpio2/direction
echo high > /sys/class/gpio/gpio13/direction
echo low > /sys/class/gpio/gpio15/direction
```
