[ $# -eq 0 ] && { echo "Usage: $0 [carname] [ipaddress]"; exit 1; }

echo "Copying..."
scp ./cars/$1/main.go pi@$2:/home/pi/gophercar/cars/$1/main.go
ssh pi@$2 ARG1=$1 ARG2=$2 'bash -s' <<'ENDSSH'
echo "Compiling..."
export PATH=$PATH:/usr/local/go/bin
GOARCH=arm GOOS=linux go build -o ./gophercar/build/$ARG1 ./gophercar/cars/$ARG1/main.go
echo "Running..."
./gophercar/build/$ARG1
ENDSSH
