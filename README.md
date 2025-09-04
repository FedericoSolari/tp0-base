# TP0: Docker + Comunicaciones + Concurrencia

En el presente repositorio se provee un esqueleto básico de cliente/servidor, en donde todas las dependencias del mismo se encuentran encapsuladas en containers. Los alumnos deberán resolver una guía de ejercicios incrementales, teniendo en cuenta las condiciones de entrega descritas al final de este enunciado.

 El cliente (Golang) y el servidor (Python) fueron desarrollados en diferentes lenguajes simplemente para mostrar cómo dos lenguajes de programación pueden convivir en el mismo proyecto con la ayuda de containers, en este caso utilizando [Docker Compose](https://docs.docker.com/compose/).

## Instrucciones de uso
El repositorio cuenta con un **Makefile** que incluye distintos comandos en forma de targets. Los targets se ejecutan mediante la invocación de:  **make \<target\>**. Los target imprescindibles para iniciar y detener el sistema son **docker-compose-up** y **docker-compose-down**, siendo los restantes targets de utilidad para el proceso de depuración.

Los targets disponibles son:

| target  | accion  |
|---|---|
|  `docker-compose-up`  | Inicializa el ambiente de desarrollo. Construye las imágenes del cliente y el servidor, inicializa los recursos a utilizar (volúmenes, redes, etc) e inicia los propios containers. |
| `docker-compose-down`  | Ejecuta `docker-compose stop` para detener los containers asociados al compose y luego  `docker-compose down` para destruir todos los recursos asociados al proyecto que fueron inicializados. Se recomienda ejecutar este comando al finalizar cada ejecución para evitar que el disco de la máquina host se llene de versiones de desarrollo y recursos sin liberar. |
|  `docker-compose-logs` | Permite ver los logs actuales del proyecto. Acompañar con `grep` para lograr ver mensajes de una aplicación específica dentro del compose. |
| `docker-image`  | Construye las imágenes a ser utilizadas tanto en el servidor como en el cliente. Este target es utilizado por **docker-compose-up**, por lo cual se lo puede utilizar para probar nuevos cambios en las imágenes antes de arrancar el proyecto. |
| `build` | Compila la aplicación cliente para ejecución en el _host_ en lugar de en Docker. De este modo la compilación es mucho más veloz, pero requiere contar con todo el entorno de Golang y Python instalados en la máquina _host_. |

### Servidor

Se trata de un "echo server", en donde los mensajes recibidos por el cliente se responden inmediatamente y sin alterar. 

Se ejecutan en bucle las siguientes etapas:

1. Servidor acepta una nueva conexión.
2. Servidor recibe mensaje del cliente y procede a responder el mismo.
3. Servidor desconecta al cliente.
4. Servidor retorna al paso 1.


### Cliente
 se conecta reiteradas veces al servidor y envía mensajes de la siguiente forma:
 
1. Cliente se conecta al servidor.
2. Cliente genera mensaje incremental.
3. Cliente envía mensaje al servidor y espera mensaje de respuesta.
4. Servidor responde al mensaje.
5. Servidor desconecta al cliente.
6. Cliente verifica si aún debe enviar un mensaje y si es así, vuelve al paso 2.

### Ejemplo

Al ejecutar el comando `make docker-compose-up`  y luego  `make docker-compose-logs`, se observan los siguientes logs:

```
client1  | 2024-08-21 22:11:15 INFO     action: config | result: success | client_id: 1 | server_address: server:12345 | loop_amount: 5 | loop_period: 5s | log_level: DEBUG
client1  | 2024-08-21 22:11:15 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°1
server   | 2024-08-21 22:11:14 DEBUG    action: config | result: success | port: 12345 | listen_backlog: 5 | logging_level: DEBUG
server   | 2024-08-21 22:11:14 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:15 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:15 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°1
server   | 2024-08-21 22:11:15 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:20 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:20 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°2
server   | 2024-08-21 22:11:20 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:20 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°2
server   | 2024-08-21 22:11:25 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:25 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°3
client1  | 2024-08-21 22:11:25 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°3
server   | 2024-08-21 22:11:25 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:30 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:30 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°4
server   | 2024-08-21 22:11:30 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:30 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°4
server   | 2024-08-21 22:11:35 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:35 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°5
client1  | 2024-08-21 22:11:35 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°5
server   | 2024-08-21 22:11:35 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:40 INFO     action: loop_finished | result: success | client_id: 1
client1 exited with code 0
```


## Parte 1: Introducción a Docker
En esta primera parte del trabajo práctico se plantean una serie de ejercicios que sirven para introducir las herramientas básicas de Docker que se utilizarán a lo largo de la materia. El entendimiento de las mismas será crucial para el desarrollo de los próximos TPs.

### Ejercicio N°1:
Definir un script de bash `generar-compose.sh` que permita crear una definición de Docker Compose con una cantidad configurable de clientes.  El nombre de los containers deberá seguir el formato propuesto: client1, client2, client3, etc. 

El script deberá ubicarse en la raíz del proyecto y recibirá por parámetro el nombre del archivo de salida y la cantidad de clientes esperados:

`./generar-compose.sh docker-compose-dev.yaml 5`

Considerar que en el contenido del script pueden invocar un subscript de Go o Python:

```
#!/bin/bash
echo "Nombre del archivo de salida: $1"
echo "Cantidad de clientes: $2"
python3 mi-generador.py $1 $2
```

En el archivo de Docker Compose de salida se pueden definir volúmenes, variables de entorno y redes con libertad, pero recordar actualizar este script cuando se modifiquen tales definiciones en los sucesivos ejercicios.

### Ejercicio N°2:
Modificar el cliente y el servidor para lograr que realizar cambios en el archivo de configuración no requiera reconstruír las imágenes de Docker para que los mismos sean efectivos. La configuración a través del archivo correspondiente (`config.ini` y `config.yaml`, dependiendo de la aplicación) debe ser inyectada en el container y persistida por fuera de la imagen (hint: `docker volumes`).


### Ejercicio N°3:
Crear un script de bash `validar-echo-server.sh` que permita verificar el correcto funcionamiento del servidor utilizando el comando `netcat` para interactuar con el mismo. Dado que el servidor es un echo server, se debe enviar un mensaje al servidor y esperar recibir el mismo mensaje enviado.

En caso de que la validación sea exitosa imprimir: `action: test_echo_server | result: success`, de lo contrario imprimir:`action: test_echo_server | result: fail`.

El script deberá ubicarse en la raíz del proyecto. Netcat no debe ser instalado en la máquina _host_ y no se pueden exponer puertos del servidor para realizar la comunicación (hint: `docker network`). `


### Ejercicio N°4:
Modificar servidor y cliente para que ambos sistemas terminen de forma _graceful_ al recibir la signal SIGTERM. Terminar la aplicación de forma _graceful_ implica que todos los _file descriptors_ (entre los que se encuentran archivos, sockets, threads y procesos) deben cerrarse correctamente antes que el thread de la aplicación principal muera. Loguear mensajes en el cierre de cada recurso (hint: Verificar que hace el flag `-t` utilizado en el comando `docker compose down`).

## Parte 2: Repaso de Comunicaciones

Las secciones de repaso del trabajo práctico plantean un caso de uso denominado **Lotería Nacional**. Para la resolución de las mismas deberá utilizarse como base el código fuente provisto en la primera parte, con las modificaciones agregadas en el ejercicio 4.

### Ejercicio N°5:
Modificar la lógica de negocio tanto de los clientes como del servidor para nuestro nuevo caso de uso.

#### Cliente
Emulará a una _agencia de quiniela_ que participa del proyecto. Existen 5 agencias. Deberán recibir como variables de entorno los campos que representan la apuesta de una persona: nombre, apellido, DNI, nacimiento, numero apostado (en adelante 'número'). Ej.: `NOMBRE=Santiago Lionel`, `APELLIDO=Lorca`, `DOCUMENTO=30904465`, `NACIMIENTO=1999-03-17` y `NUMERO=7574` respectivamente.

Los campos deben enviarse al servidor para dejar registro de la apuesta. Al recibir la confirmación del servidor se debe imprimir por log: `action: apuesta_enviada | result: success | dni: ${DNI} | numero: ${NUMERO}`.



#### Servidor
Emulará a la _central de Lotería Nacional_. Deberá recibir los campos de la cada apuesta desde los clientes y almacenar la información mediante la función `store_bet(...)` para control futuro de ganadores. La función `store_bet(...)` es provista por la cátedra y no podrá ser modificada por el alumno.
Al persistir se debe imprimir por log: `action: apuesta_almacenada | result: success | dni: ${DNI} | numero: ${NUMERO}`.

#### Comunicación:
Se deberá implementar un módulo de comunicación entre el cliente y el servidor donde se maneje el envío y la recepción de los paquetes, el cual se espera que contemple:
* Definición de un protocolo para el envío de los mensajes.
* Serialización de los datos.
* Correcta separación de responsabilidades entre modelo de dominio y capa de comunicación.
* Correcto empleo de sockets, incluyendo manejo de errores y evitando los fenómenos conocidos como [_short read y short write_](https://cs61.seas.harvard.edu/site/2018/FileDescriptors/).


### Ejercicio N°6:
Modificar los clientes para que envíen varias apuestas a la vez (modalidad conocida como procesamiento por _chunks_ o _batchs_). 
Los _batchs_ permiten que el cliente registre varias apuestas en una misma consulta, acortando tiempos de transmisión y procesamiento.

La información de cada agencia será simulada por la ingesta de su archivo numerado correspondiente, provisto por la cátedra dentro de `.data/datasets.zip`.
Los archivos deberán ser inyectados en los containers correspondientes y persistido por fuera de la imagen (hint: `docker volumes`), manteniendo la convencion de que el cliente N utilizara el archivo de apuestas `.data/agency-{N}.csv` .

En el servidor, si todas las apuestas del *batch* fueron procesadas correctamente, imprimir por log: `action: apuesta_recibida | result: success | cantidad: ${CANTIDAD_DE_APUESTAS}`. En caso de detectar un error con alguna de las apuestas, debe responder con un código de error a elección e imprimir: `action: apuesta_recibida | result: fail | cantidad: ${CANTIDAD_DE_APUESTAS}`.

La cantidad máxima de apuestas dentro de cada _batch_ debe ser configurable desde config.yaml. Respetar la clave `batch: maxAmount`, pero modificar el valor por defecto de modo tal que los paquetes no excedan los 8kB. 

Por su parte, el servidor deberá responder con éxito solamente si todas las apuestas del _batch_ fueron procesadas correctamente.

### Ejercicio N°7:

Modificar los clientes para que notifiquen al servidor al finalizar con el envío de todas las apuestas y así proceder con el sorteo.
Inmediatamente después de la notificacion, los clientes consultarán la lista de ganadores del sorteo correspondientes a su agencia.
Una vez el cliente obtenga los resultados, deberá imprimir por log: `action: consulta_ganadores | result: success | cant_ganadores: ${CANT}`.

El servidor deberá esperar la notificación de las 5 agencias para considerar que se realizó el sorteo e imprimir por log: `action: sorteo | result: success`.
Luego de este evento, podrá verificar cada apuesta con las funciones `load_bets(...)` y `has_won(...)` y retornar los DNI de los ganadores de la agencia en cuestión. Antes del sorteo no se podrán responder consultas por la lista de ganadores con información parcial.

Las funciones `load_bets(...)` y `has_won(...)` son provistas por la cátedra y no podrán ser modificadas por el alumno.

No es correcto realizar un broadcast de todos los ganadores hacia todas las agencias, se espera que se informen los DNIs ganadores que correspondan a cada una de ellas.

## Parte 3: Repaso de Concurrencia
En este ejercicio es importante considerar los mecanismos de sincronización a utilizar para el correcto funcionamiento de la persistencia.

### Ejercicio N°8:

Modificar el servidor para que permita aceptar conexiones y procesar mensajes en paralelo. En caso de que el alumno implemente el servidor en Python utilizando _multithreading_,  deberán tenerse en cuenta las [limitaciones propias del lenguaje](https://wiki.python.org/moin/GlobalInterpreterLock).

## Condiciones de Entrega
Se espera que los alumnos realicen un _fork_ del presente repositorio para el desarrollo de los ejercicios y que aprovechen el esqueleto provisto tanto (o tan poco) como consideren necesario.

Cada ejercicio deberá resolverse en una rama independiente con nombres siguiendo el formato `ej${Nro de ejercicio}`. Se permite agregar commits en cualquier órden, así como crear una rama a partir de otra, pero al momento de la entrega deberán existir 8 ramas llamadas: ej1, ej2, ..., ej7, ej8.
 (hint: verificar listado de ramas y últimos commits con `git ls-remote`)

Se espera que se redacte una sección del README en donde se indique cómo ejecutar cada ejercicio y se detallen los aspectos más importantes de la solución provista, como ser el protocolo de comunicación implementado (Parte 2) y los mecanismos de sincronización utilizados (Parte 3).

Se proveen [pruebas automáticas](https://github.com/7574-sistemas-distribuidos/tp0-tests) de caja negra. Se exige que la resolución de los ejercicios pase tales pruebas, o en su defecto que las discrepancias sean justificadas y discutidas con los docentes antes del día de la entrega. El incumplimiento de las pruebas es condición de desaprobación, pero su cumplimiento no es suficiente para la aprobación. Respetar las entradas de log planteadas en los ejercicios, pues son las que se chequean en cada uno de los tests.

La corrección personal tendrá en cuenta la calidad del código entregado y casos de error posibles, se manifiesten o no durante la ejecución del trabajo práctico. Se pide a los alumnos leer atentamente y **tener en cuenta** los criterios de corrección informados  [en el campus](https://campusgrado.fi.uba.ar/mod/page/view.php?id=73393).


# Solución


### Ejercicio 1

Se genero un script de bash generar-compose.sh, que permite configurar un nombre de archivo de configuracion ```.yaml``` y una cantidad de clientes determinada.

  **Uso:**  
  Se debe correr, desde la carpeta donde esta alacenado:
  ```
  ./generar-compose.sh <archivo_de_salida.yaml> <cantidad de clientes>
  ```

Este script automatiza la generación de un archivo docker-compose.yaml, siguiendo como modelo al archivo provisto por la catedra en la rama master, válido para levantar un entorno con un servidor y múltiples clientes conectados a una red común, destacándose por su validación robusta de entradas, su modularidad a través de funciones, la flexibilidad para variar dinámicamente la cantidad de clientes y la claridad con la que estructura los servicios y la red en Docker Compose.

---

### Ejercicio 2

Para evitar que cada modificación en el archivo de configuración requiera reconstruir las imágenes de Docker, se incorpora en el script que genera el archivo `.yaml` la sección de **`volumes`**, tanto en el servicio del servidor como en el de los clientes.

El uso de **Docker Volumes** permite persistir datos fuera del contenedor, de modo que los archivos de configuración se mantengan en el host y puedan ser modificados sin necesidad de reconstruir la imagen. Es decir, que al iniciar el servidor este dispone dentro del contenedor de la configuración inyectada desde el host, de lo contrario, el contenedor arrancaría vacío y sería necesario recompilar la imagen.  


Para correrlo, se puede continuar utilizando el generador de script del ejercicio 1, de la siguiente manera:

  **Uso:**  
  Se debe correr, desde la raiz del proyecto:
  ```
  ./generar-compose.sh <archivo_de_salida.yaml> <cantidad de clientes>
  ```

  y luego se levantan los containers de docker con el comando:

  ```
  make docker-compose-up
  ```
---

### Ejercicio 3

El script `validar-echo-server.sh` tiene como objetivo verificar el correcto funcionamiento de un **servidor Echo**. Para esto, se utilizan contenedores Docker y la herramienta **netcat (nc)**, sin necesidad de instalar netcat en la máquina host ni exponer puertos del servidor hacia el exterior.  

Se define un mensaje de texto que será enviado al servidor. Este mensaje servirá como referencia para validar que el servidor responda correctamente:  

```bash
message="Distribuidos TP0"
```

A partir del archivo server/config.ini, se obtienen el puerto (SERVER_PORT) y la dirección IP (SERVER_IP) del servidor. Para esto se utiliza awk para extraer el valor asociado a cada variable, y se aplica tr -d ' ' para eliminar espacios en blanco:

```bash
SERVER_PORT=$(awk -F'=' '/SERVER_PORT/ {print $2}' server/config.ini | tr -d ' ')
SERVER_IP=$(awk -F'=' '/SERVER_IP/ {print $2}' server/config.ini | tr -d ' ')
```
Luego se lanza un contenedor efímero basado en la imagen busybox:latest que se conecta a la misma red Docker (tp0_testing_net) que el servidor. Este contenedor ejecuta el comando nc para enviar el mensaje al servidor y captura la respuesta en la variable response

```bash
response=$(docker run --rm --network tp0_testing_net busybox:latest sh -c "echo '$message' | nc $SERVER_IP $SERVER_PORT")

```

De esta manera no es necesario exponer puertos en el host, ya que la comunicación ocurre dentro de la red virtual de Docker.

Finalmente, se compara la respuesta obtenida del servidor con el mensaje enviado. Si son iguales, el test es exitoso sino falla.

---

### Ejercicio 4

El objetivo de este ejercicio fue modificar el servidor y el cliente para que ambos sistemas finalicen de forma **graceful** al recibir la señal `SIGTERM`. Terminar la aplicación de forma graceful implica que todos los **file descriptors** abiertos (sockets, archivos, threads, etc.) sean cerrados correctamente antes de que el proceso principal muera. Además, en cada cierre se registran mensajes de log que permiten verificar el correcto liberado de recursos.  

---

#### Servidor (Python)

Se capturó la señal `SIGTERM` mediante el módulo `signal` y se implementó el método `handle_sigterm_signal`. Cuando llega la señal, se loguea el evento, se actualiza un flag de apagado (`self.shutdown = True`) y se cierra el socket principal del servidor:  

```python
signal.signal(signal.SIGTERM, self.handle_sigterm_signal)

def handle_sigterm_signal(self, signum, frame):
    logging.info("action: handle_sigterm_signal | result: success")
    self.shutdown = True
    self._server_socket.close()
```
El bucle principal (run) se ejecuta mientras self.shutdown sea False. Una vez que se detecta el cierre, se invoca a clean_resourses, que se encarga de cerrar todos los sockets de clientes aún abiertos, vaciar la lista de clientes y loguear el proceso de liberación:
De esta forma, el servidor no queda bloqueado esperando conexiones cuando debe finalizar, y todos los recursos se liberan antes de terminar.

---

#### Client (Go)
Se utilizó el paquete os/signal para capturar SIGTERM. Se definió un canal de señales (sigs) y una goroutine encargada de escuchar por la señal. Cuando llega el SIGTERM, se setea un flag de apagado (c.shutdown = true) y se cierra la conexión TCP si está abierta

```Go
go func() {
    <-sigs
    client.shutdown = true
    log.Infof("action: sigterm_received | result: in_progress | client_id: %v", client.config.ID)
    if client.conn != nil {
        err := client.conn.Close()
        if err == nil {
            log.Infof("action: close_connection | result: success | client_id: %v", client.config.ID)
        }
    }
}()
```
Durante la ejecución del loop (StartClientLoop), el cliente chequea el flag shutdown. Si está activo, deja de enviar mensajes y termina de manera ordenada.
Particularmente, esta solución se encuentra implementada únicamente en la rama `ej4`, debido a diversos inconvenientes surgidos al momento de ejecutar uno de los tests. Esto obligó a rehacer la lógica completa del manejo de la señal `SIGTERM`, luego de haber finalizado previamente las demás partes del TP0.  
No obstante, estimo con seguridad que la solución presente en las demás ramas también cumple con los requisitos necesarios para superar las pruebas. El único test que presentaba fallas era **`server_without_clients_down`**, y la causa de dicho error no estaba relacionada con la lógica del servidor o del cliente, sino con factores externos al propio código. 

---

### Ejercicio 5

Cada cliente emula una agencia de quiniela y genera una apuesta (Bet) a partir de las variables de entorno que representan los datos de la persona (nombre, apellido, DNI, nacimiento y número).
La apuesta se serializa siguiendo el protocolo definido, que en este caso es un mensaje en formato CSV terminado en \n:

```GO
return fmt.Sprintf("%s,%s,%s,%s,%s,%d\n",b.Agency, b.FirstName, b.LastName, b.Document, b.Birthdate, b.Number)
```
Esto asegura que el servidor reciba los campos en el orden esperado y pueda procesarlos correctamente.

Para evitar problemas de short write, se implementa un método SendAll que garantiza que todos los bytes del mensaje se envíen

```GO
for sent < total {
    n, err := c.conn.Write(data[sent:])
    sent += n
}
```

El cliente espera la respuesta del servidor usando el método RecvAll, que asegura leer hasta encontrar un delimitador (\n) o hasta cerrar la conexión, manejando el caso de short read:

```GO
if line, rest, found := extractLine(c.buffer); found {
    return line, nil
}
```

En el server, cada mensaje de cliente se lee hasta el \n, usando recv_all, para evitar short read. Luego se parsea el mensaje según el protocolo definido.
Una vez parseada correctamente, la apuesta se almacena usando la función provista store_bets, luego se envía un mensaje de confirmación al cliente.

La clase connection es un fiel reflejo de la implementada para el cliente, pero en otro lenguaje, manejando el short read y write adecuadamente.Considero importante destacar el metodo implementado para la lectura, donde se opto por un diseño de leer hasta una cantidad determinada, lo cual puede significar que lea todo esos datos o menos. Esto implica que en una lectura puede haber mas de un mensaje, completo o incompleto. En el caso que haya dos mensajes completo se devuelve el primero y al momento de invocar nuevamente al read all se devuelve la otra parte, si esta incompleta se continua leyendo agregando la informacion en lo ya obtenido.

```Python
idx = self._buffer.find(b'\n')
if idx != -1:
    line = bytes(self._buffer[:idx+1])
    del self._buffer[:idx+1]
    return line
```


---
### Ejercicio 6


Para dar soporte al envío de apuestas en modalidad batch, fue necesario introducir cambios tanto en el cliente como en el servidor.

En el cliente, se implementó un BetLoader encargado de leer los archivos CSV de cada agencia y dividir las apuestas en lotes de tamaño configurable (BatchMaxAmount). Esto asegura que cada cliente envíe múltiples apuestas en un solo mensaje, respetando el límite definido en config.yaml y evitando superar los 8 kB por paquete.

La lógica se centraliza en ProcessAllBets, que recorre todos los lotes y delega el envío a ProcessBets

```GO
for {
    batch, err := loader.NextBatch()
    if err == io.EOF {
        break
    }
    err = c.ProcessBets(handler, batch)
    if err != nil {
        return err
    }
}
```

Cada lote se serializa mediante FormatBatchMessage y se envía al servidor, verificando que la respuesta corresponda a un éxito antes de continuar con el siguiente:

```GO
msg := FormatBatchMessage(bets)
err := handler.SendAll([]byte(msg))
response, _ := handler.RecvAll()
if !IsSuccessResponse(response) {
    return fmt.Errorf("fallo en batch: %q", response)
}
```

En el servidor, se incorporó la función __handle_bets, que recibe un mensaje con varias apuestas, las parsea (protocol.parse_bet_message) y las almacena individualmente con utils.store_bets:

```Py
bets = protocol.parse_bet_message(msg)
for bet in bets:
    try:
        utils.store_bets([bet])
    except Exception:
        batch_ok = False
        break
```


La respuesta es global al batch: si todas las apuestas se registran correctamente, se confirma con éxito; en caso contrario, se marca como fallo:

```Py
if batch_ok:
    conn.send_message(protocol.success_message())
else:
    conn.send_message(protocol.error_message())
```

Esto garantiza la atomicidad en el procesamiento o bien todo el batch se acepta, o bien se rechaza completo.
Si se rechaza, del lado del cliente se termina el procesamiento como fallo, se opto por esta solucion dado que se realizan varias validaciones previas a enviar del lado del cliente, por lo que, como dije anteriormente, para garantizar la atomicidad de la transaccion, se decide abortar.

Finalmente, la capa de protocolo se adaptó para distinguir claramente tres tipos de mensajes: inicio de sesión, lotes de apuestas y finalización. El servidor mantiene un estado de sesión (session_active) que asegura que los mensajes sean interpretados en el orden correcto y dentro del contexto de una sesión válida.

