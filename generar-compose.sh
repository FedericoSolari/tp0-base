#!/bin/bash

function return_error() {
    echo "Error: $1"
    exit 1
}

# Valido que tenga los parametros necesarios
# Valido que vengan en el formato pedido
function input_control() {
    if [ -z "$file" ] || [ -z "$clients" ]; then
        return_error "Debe proporcionar el nombre del archivo y la cantidad de clientes. Ej: ./generar-compose.sh docker-compose-dev.yaml 5"
    fi

    if ! [[ "$clients" =~ ^[0-9]+$ ]]; then
        return_error "La cantidad de clientes debe ser un número entero"
    fi

    if [[ "$file" != *.yaml ]]; then
        return_error "El archivo de salida debe terminar con .yaml"
    fi
}


function write_line() {
    echo "$1" >> "$file"
}

function configurate_server() {
    # Piso el archivo
    echo "name: tp0" > "$file"

    write_line "services:"
    write_line "  server:"
    write_line "    container_name: server"
    write_line "    image: server:latest"
    write_line "    entrypoint: python3 /main.py"
    write_line "    environment:"
    write_line "      - PYTHONUNBUFFERED=1"
    write_line "    networks:"
    write_line "      - testing_net"
    write_line "    volumes:"
    write_line "      - ./server/config.ini:/server/config.ini"
}

function configurate_client() {
    for ((i=1; i<=$clients; i++))
    do
        write_line "  client$i:"
        write_line "    container_name: client$i"
        write_line "    image: client:latest"
        write_line "    entrypoint: /client"
        write_line "    environment:"
        write_line "      - CLI_ID=$i"
        write_line "      - NOMBRE=Nombre$i"
        write_line "      - APELLIDO=Apellido$i"
        write_line "      - DOCUMENTO=4000000$i"
        write_line "      - NACIMIENTO=1990-01-0$i"
        write_line "      - NUMERO=75$i"
        write_line "    networks:"
        write_line "      - testing_net"
        write_line "    volumes:"
        write_line "      - ./client/config.yaml:/config.yaml"
        write_line "    depends_on:"
        write_line "      - server"
    done
}

function configurate_network() {
    write_line ""
    write_line "networks:"
    write_line "  testing_net:"
    write_line "    ipam:"
    write_line "      driver: default"
    write_line "      config:"
    write_line "        - subnet: 172.25.125.0/24"
}

file=$1
clients=$2

input_control
configurate_server
configurate_client
configurate_network
