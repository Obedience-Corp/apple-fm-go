set working-directory := ".."

# Beta source-release gates. Native compatibility is a separate stable-release gate.
check:
    just lint
    just test
    just check cross-build
    go mod verify
