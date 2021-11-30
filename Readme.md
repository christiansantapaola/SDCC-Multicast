# Sistemi distribuiti e cloud computing: Multicast applicativo
Questo è il codice sorgente del progetto B1: Multicast Applicativo per il corso
di SDCC dell'università di Roma Tor Vergata.

# Installazione:
fare il clone del progetto:
``` sh
$ git clone https://github.com/christiansantapaola/SDCC-Multicast
```
entrare nella cartella creata ed entrare nella sottocartella Docker, qui dentro
ci saranno i DockerFile che descrivono i container del name service e del client.
``` sh
$ cd SDCC-Multicast/docker
```
## Name Service
entrare dentro la cartella del name service e eseguire la build del container:
``` sh
$ cd sdcc-name-service
$ docker-compose build
$ docker-compose up
```
## Client
Queste istruzioni permettono di creare il container per un client, creare tanti
container e ripetere queste istruzione per il numero di utenti client
desiderati.
Aprire un altra shell ed entrare nella cartella SDCC-Multicast/docker/sdcc e
fare la build del client
``` sh
$ cd SDCC-Multicast/docker/sdcc
$ docker build -t sdcc .
``` 
completata la build si fa partire il client come:
``` sh
$ docker run --network sdcc-network -it sdcc
```
questo comando farà partire una shell all'interno del container con il progetto
al suo interno.
Prima di far partire il programma ci serve l'indirizzo IP del client e del name
server all'interno della rete docker, quindi su una terza shell chiamiamo:
``` sh
$ docker network inspect sdcc-network
```
nell'output di questo comando dovremmo trovare gli indirizzi ip del name service
e del nostro client appena creato.
torniamo alla shell del client ed eseguiamo:
```
# sdcc_gen_config
```
per generare la config nel file 'sdcc_config.yml', lo apriamo è troviamo questo
file:
``` yaml
name_server_address: 127.0.0.1:2080
user_id: ""
user_ip: 127.0.0.1
user_port: 2079
secure: false
verbose: false
timeout: 5s
```
Inserire gli indirizzi del name service e il nostro presi dal comando docker
rispettivamente in name_server_address e user_ip, non modificare le porte.

eseguire poi
``` sh 
sdcc_create_user
sdcc_create_group -group <nome-gruppo>
sdcc_join_group -group <nome-gruppo>
```
Per registrare l'utente, il gruppo e fare il join.
eseguire poi a seconda del tipo di comunicazione che si vuole `sdcc_seq`,
`sdcc_lc`, `sdcc_vc` per multicast con sequencer (seq), multicast con clock
scalare (lc), multicast con clock vettoriale (vc).
