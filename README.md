# dynsim

physics simulation in your terminal. that's it.

## install

```bash
go install github.com/san-kum/dynsim/cmd/dynsim@latest
```

or clone and build:

```bash
git clone https://github.com/san-kum/dynsim
cd dynsim
go build -o dynsim cmd/dynsim/main.go
```

## run it

```bash
./dynsim
```

you get a menu. pick a model. tweak params. watch physics happen.

## models

| model           | what it does                                        |
| --------------- | --------------------------------------------------- |
| pendulum        | swings back and forth. simple harmonic motion.      |
| double_pendulum | two pendulums chained. goes chaotic.                |
| cartpole        | balance a stick on a cart. classic control problem. |
| spring_mass     | bouncy mass on a spring.                            |
| drone           | 2d quadrotor. tries to hover.                       |
| nbody           | gravitational attraction between bodies.            |

## cli

don't want the tui? use the cli:

```bash
# run a simulation
./dynsim run pendulum --theta 0.5 --time 10

# use different integrator
./dynsim run pendulum --integrator rk45

# live visualization
./dynsim live pendulum --theta 1.0

# export to csv
./dynsim export-csv pendulum --time 5 > data.csv

# compare integrators
./dynsim compare pendulum euler rk4 rk45
```

## presets

quick demos:

```bash
./dynsim run pendulum --preset small    # gentle swing
./dynsim run pendulum --preset large    # big swing
./dynsim run double_pendulum --preset chaos  # butterfly effect
./dynsim run cartpole --preset balance  # stays upright
./dynsim run drone --preset hover       # hovers at y=5
```

## integrators

| name   | accuracy   | speed  | use when                        |
| ------ | ---------- | ------ | ------------------------------- |
| euler  | low        | fast   | you don't care about accuracy   |
| rk4    | high       | medium | default, works for most things  |
| rk45   | adaptive   | varies | stiff systems, long simulations |
| verlet | symplectic | fast   | energy conservation matters     |

## config files

yaml configs work too:

```yaml
model: pendulum
integrator: rk4
dt: 0.01
duration: 10.0
init_state:
  theta: 0.5
  omega: 0.0
```

```bash
./dynsim run pendulum --config experiment.yaml
```

## how it works

1. you pick a model (defines the physics equations)
2. you pick an integrator (solves the differential equations)
3. simulation steps through time
4. state updates based on derivatives
5. you see pretty things on screen

the core loop:

```
for t < duration:
    u = controller.compute(state, t)
    state = integrator.step(dynamics, state, u, t, dt)
    t += dt
```

## add your own model

implement the `Dynamics` interface:

```go
type Dynamics interface {
    Derivative(x State, u Control, t float64) State
    StateDim() int
    ControlDim() int
}
```

register it in `internal/experiment/registry.go`:

```go
r.models["mymodel"] = func() sim.Dynamics { return NewMyModel() }
```

done.

## keyboard shortcuts (tui)

| key   | what               |
| ----- | ------------------ |
| j/k   | move up/down       |
| enter | select             |
| s     | start simulation   |
| space | pause/resume       |
| +/-   | speed up/slow down |
| r     | reset              |
| c     | back to config     |
| q     | quit               |

## why

wanted to understand physics simulations. built this to play around with different models and integrators. turns out watching things in _motion_ is oddly satisfying.

## license

mit. do whatever you want.
