tab="--tab"
cmd="bash -c 'go run *.go';bash"
foo=""

for i in 1 2 ... n; do
      foo+=($tab -e "$cmd"+$1)         
done

gnome-terminal "${foo[@]}"

exit 0
