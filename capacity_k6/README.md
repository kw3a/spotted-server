# Precondiciones y tests de capacidad

Hay dos ciclos separados: un dataset base reutilizable y los datos temporales de
cada ejecucion. Todos los VUs son aplicantes y registran su propia cuenta por HTTP.
El numero de VUs no cambia el seed ni exige editar archivos de usuarios o quizzes.

## 1. Crear las precondiciones

Todos los comandos siguientes se ejecutan dentro de `capacity_k6`:

```bash
cd capacity_k6
cp .env.example .env
```

Editar `.env` con `DATABASE_URL` y `BASE_URL` del servidor objetivo. La base debe
ser la misma que usa esa API. Este archivo es independiente del `.env` de la raiz;
el lanzador no carga la configuracion del servidor desde la raiz.
El CLI lee el archivo con `godotenv.Read`: no se necesita `export`, `source`
ni prefijos `env` en la terminal. El archivo es la fuente de las conexiones;
valores previamente exportados no lo sobrescriben. Solo se admiten
`DATABASE_URL` y `BASE_URL` en dotenv. Las opciones de carga se pasan como flags
al comando `run`; si el archivo contiene otras claves, se informa cuales deben
moverse a flags.

La carga requiere Go y k6, BASE_URL y el manifest ya generado. El seeder es un
comando independiente: requiere MySQL accesible y ambas migraciones aplicadas
(esquema y triggers). Si ya ejecutaste el seeder, no lo repitas para lanzar carga:

```bash
./seed.sh
```

El script compila el CLI Go de `fixtures` en `.bin`, carga `DATABASE_URL` de
`.env` mediante godotenv y crea:

| Entidad | Cantidad |
|---|---:|
| Usuarios propietarios de las companias | 10 |
| Companias | 10 |
| Ofertas activas | 10 |
| Pruebas algorítmicas / quizzes | 10 |
| Problemas: 3 por quiz | 30 |
| Ejemplos publicos: 2 por problema | 60 |
| Casos secretos: 5 por problema | 150 |

Habilita C++ (Judge0 ID 54) y Java (ID 62) para cada quiz y crea sus registros
si no existen. Son los mismos IDs usados por el seeder del servidor. Judge0 del
dispositivo externo debe tener esos lenguajes habilitados; crear las filas en
MySQL no instala compiladores en Judge0.

Los 30 problemas alternan el lenguaje por quiz y posicion: 15 usan C++ y 15 Java.
Cada quiz contiene ambos lenguajes; cada VU los utiliza al resolver los tres
problemas. Las variantes correcta, parcial y con error de sintaxis de un problema
usan su mismo lenguaje. Se exige cobertura y se separan metricas por `language`
(`cpp`/`java`) y `solution_type`. Los fuentes estan en `fixtures/solutions/`,
incluidos en el binario del generador; Java utiliza la clase publica `Main`.
Cada quiz dura 180 minutos y contiene estos algoritmos:

- Subsecuencia creciente estricta mas larga: busqueda binaria sobre los finales
  de las subsecuencias, O(n log n), hasta 10000 elementos.
- Mochila 0/1: programacion dinamica O(n W), hasta 300 objetos y capacidad 2000.
- Camino minimo ponderado: Dijkstra con heap, en grafos dirigidos con pesos no
  negativos, ciclos, aristas paralelas y vertices inalcanzables.

Los enunciados especifican formato de entrada/salida, restricciones y complejidad.
Los casos secretos combinan contraejemplos pequenos y entradas grandes: una
secuencia de 10000 elementos, 300 objetos con capacidad 2000 y un grafo de
1500 vertices con 4495 aristas. Las entradas caben en los campos TEXT actuales;
no es necesario cambiar el esquema. Cada quiz reutiliza estos tres problemas
con IDs propios; el contenido es determinista y no depende del numero de VUs.

Los cinco casos secretos de cada problema son distintos de sus dos ejemplos.
Estan en `test_case`, separado de `example`: no se incluyen en el archivo que
consume k6 ni se muestran como ejemplos.

El seed usa UUIDs deterministas, transacciones y comprobaciones de relaciones y
contenido. Repetirlo no duplica el dataset ni borra datos existentes. Si una
precondicion existente fue modificada de manera incompatible, devuelve un error.
Los diez usuarios base se llaman `capacity_base_01` ... `capacity_base_10` y tienen
contrasenas aleatorias hasheadas; solo proporcionan la propiedad de las companias.
Los VUs no usan esas cuentas ni necesitan conocer sus credenciales.

Se escribe `fixtures.json` con los IDs de las ofertas/quizzes,
los tres problemas y sus soluciones C++ y Java. No contiene credenciales, entradas
ni salidas de los casos secretos. No depende del numero de VUs.

## 2. Ejecutar la carga por HTTP

Primero, un smoke: cada VU completa un recorrido y resuelve los tres problemas.
El lanzador toma BASE_URL de dotenv y las opciones de carga de sus flags.
Con `--think-scale 1` (predeterminado), cada envio de codigo tiene una pausa
previa de escritura de 60 a 120 segundos. El VU tambien espera el resultado del
intento anterior y consulta su puntuacion antes de preparar el siguiente.

```bash
./run.sh run --smoke --vus 1 --think-scale 0
```

Prueba con rampas y pausas de lectura/escritura:

```bash
./run.sh run --vus 110 \
  --summary-export=full.summary.json \
  --out json=full.results.json
```

Para cambiar el volumen basta con modificar `--vus`, por ejemplo 55, 110, 220.
No es necesario volver a generar precondiciones entre esas ejecuciones.

El lanzador de carga no incluye seeders ni dependencias SQL y no conecta a MySQL.
Genera un RUN_ID unico y guarda su registro en `runs/<RUN_ID>.json`. Lee el manifest
local previamente generado; k6 comprueba las ofertas mediante endpoints HTTP.
Todos los registros de cuentas, participaciones, submissions y pulsaciones se
realizan por HTTP contra BASE_URL. El servidor procesa los envios con su Judge0.
Los parametros se transmiten a k6 en un modulo temporal privado por ejecucion,
sin `-e` ni variables de entorno adicionales. El modulo no contiene DATABASE_URL
y se elimina al terminar. El registro guarda las opciones y el fin de la carga.
La limpieza se ejecuta despues mediante un comando SQL independiente.
Para otros archivos:

```bash
./seed.sh --env-file .env.test --manifest fixtures.test.json
./run.sh run --env-file .env.test --manifest fixtures.test.json --vus 220
```

Para apuntar a un dispositivo externo, configurar en `.env`, por ejemplo:

```dotenv
DATABASE_URL="usuario:contraseña@tcp(192.168.1.50:3306)/spotted-db?parseTime=true"
BASE_URL="https://spotted.ejemplo.com"
```

La API debe ser accesible por HTTPS y Judge0 debe poder entregar sus callbacks.
La conexion MySQL puede ser directa o mediante un tunel SSH; la base puede
permanecer en el dispositivo externo aunque el extremo del tunel sea local.
BASE_URL es obligatoria, debe usar HTTPS y rechaza direcciones loopback. Se ha
eliminado el fallback a localhost y la adaptacion de cookies para HTTP local.
Todos los paths relativos (`--env-file`, `--manifest`, `--run-dir` y salidas de k6)
se resuelven dentro de `capacity_k6`, incluso si se invoca el lanzador desde otro
directorio. Alli quedan `.bin/`, los fixtures, `runs/` y los resultados exportados.

DATABASE_URL solo se usa en ./seed.sh y ./cleanup.sh y debe corresponder a la
base usada por BASE_URL. ./run.sh ignora esta conexion y funciona con un dotenv
que contenga solo BASE_URL; no necesita acceso SQL ni permisos sobre MySQL.

## Recorrido de cada VU

1. Registra una cuenta con nick `ck_<RUN_ID>_<VU>_<generacion>`. Comprueba HTTP 201,
   redireccion al perfil y ambas cookies. Las visitas posteriores hacen login.
2. Busca y lee ofertas precargadas, usando pausas y lecturas sin reemplazo.
3. Participa en un quiz base y visita su pagina.
4. Lee los tres enunciados y sus dos ejemplos; envia las soluciones C++ y Java.
5. Por problema envia tres intentos: error de sintaxis, solucion parcial y solucion
   correcta. Verifica por SSE los cinco resultados esperados de cada intento,
   sin pendientes, y comprueba que el puntaje conserve el mejor resultado.
   El ultimo intento debe aceptar cinco casos y obtener **5/5**.
6. Finaliza la participacion y cierra sesion.

Cada cuenta puede participar una vez en cada uno de los diez quizzes. Al
consumirlos todos, el mismo VU registra otra cuenta y continua con los mismos
fixtures. Esto evita agotar precondiciones durante pruebas largas. En smoke se
crea exactamente una cuenta por VU; en carga prolongada puede haber mas.

No se crean companias, ofertas, quizzes ni problemas durante la carga.
Se crean cuentas, sesiones, participaciones, submissions, resultados y ventanas
de pulsaciones. Las cuentas de distintos VUs o ejecuciones nunca se comparten.

Cada problema incluye las tres variantes en el manifiesto:

| Problema | Correcta | Parcial | Error de sintaxis |
|---|---:|---:|---:|
| Subsecuencia creciente | 5/5 | 2/5: acepta valores iguales | 0/5 |
| Mochila 0/1 | 5/5 | 2/5: recorre capacidades en orden ascendente y reutiliza objetos | 0/5 |
| Camino minimo | 5/5 | 3/5: usa BFS y conserva la primera ruta descubierta | 0/5 |

Los errores de sintaxis deben producir cinco resultados `Error`; la API agrupa
los errores de compilacion/ejecucion bajo esa etiqueta. Las soluciones parciales
deben producir exactamente los aceptados y respuestas equivocadas previstos.
Estos rechazos deliberados son resultados correctos del juez y no cuentan como
fallos funcionales. Un veredicto inesperado, un timeout o una respuesta HTTP
incorrecta hace fallar la prueba. Las metricas usan `solution_type` y exigen cobertura
y resultados correctos para las tres categorias.

El manifiesto usa version 4. Para una base nueva o la version 3 con estos mismos
algoritmos, ejecutar `./seed.sh`: habilita C++/Java y escribe `fixtures.json`
con las nuevas soluciones, manteniendo problemas y casos. Las asociaciones de
lenguajes anteriores se conservan, pero la carga solo envia C++ y Java.
Para sustituir los problemas triviales de una version anterior, detener las
pruebas, limpiar sus runs y ejecutar desde este directorio:

```bash
./seed.sh --refresh-problems
```

La actualizacion cambia enunciados, limites, ejemplos y casos de los IDs
propios del dataset, dentro de una transaccion. Conserva usuarios, companias,
ofertas, quizzes y el numero de filas. Rechaza quizzes con participaciones
existentes para evitar modificar casos vinculados a resultados anteriores.
Si hay participaciones, limpiar primero sus runs; no se borra su historial para
forzar una actualizacion. El seed normal rechaza contenido incompatible en vez
de modificarlo silenciosamente. Al terminar se escribe el manifiesto nuevo.

## Limpieza

Una vez detenido k6, ejecutar ./cleanup.sh con el RUN_ID mostrado por el lanzador.
Este comando SQL independiente elimina solo las cuentas que coinciden
con el prefijo **y** el nombre de esa ejecucion, junto con:

- JWTs de esos usuarios, incluso si no pudieron cerrar sesion.
- Participaciones, submissions, resultados y ventanas de pulsaciones.
- Dependencias del perfil y companias pertenecientes a esas cuentas, si existen.

Conserva los diez usuarios base, sus companias, ofertas, quizzes, problemas,
ejemplos/casos y los datos de otras ejecuciones. La limpieza SQL es transaccional
e idempotente. Los archivos de metricas y los registros de ejecucion se conservan.
La limpieza se verifica en MySQL; los jobs internos de Judge0 no se administran
con este CLI.

El lanzador no limpia automaticamente, tampoco si k6 falla o recibe SIGINT/SIGTERM.
El comando de limpieza se imprime antes y despues de la carga. Ante SIGKILL o
apagado, recuperar el RUN_ID de `runs/<RUN_ID>.json` y ejecutar, con k6 detenido:

```bash
./cleanup.sh --env-file .env.b --run-id 0123456789abcdef
```

`run` y `cleanup` son pasos separados. El seeder tampoco se ejecuta en la carga.

## Parametros de terminal y resultados

`./run.sh run --help` muestra todos los flags. No se usan variables de entorno
para las opciones de carga ni se aceptan flags `-e`/`--env` de k6.

| Flag | Default / comportamiento |
|---|---|
| `--mode full` | Participa y resuelve los tres problemas; `browse` solo registra, autentica y lee |
| `--smoke` | Un recorrido por VU sin rampas; omitido usa rampas |
| `--vus 110` | Numero maximo de VUs |
| `--ramp 2m --hold 5m --ramp-down 30s` | Subida, periodo estable y bajada |
| `--think-scale 1` | Pausas normales; 0 las omite |
| `--p-filter 0.6 --p-apply 1 --p-finish 1` | Probabilidades; full requiere participacion y finalizacion positivas |
| `--submissions-per-problem 3` | Entre 3 y 30; termina con la solucion correcta |
| `--enable-keystrokes=false` | Desactiva pulsaciones; habilitadas por defecto en full |
| `--enable-results=true` | Obligatorio en full; resultados y pulsaciones se desactivan por defecto en browse |
| `--results-timeout 65s` | Timeout de SSE |
| `--results-p95-ms 60000` | Limite p95 de SSE y juez |
| `--grace 12m` | Tiempo para completar recorridos; default 30s en browse |
| `--summary-export b.summary.json` | Resumen JSON |
| `--out json=b.results.json` | Puntos de metricas; permite repetir `--out` |
| `--quiet --no-color` | Presentacion de k6 |

Los booleanos se activan con `--smoke`, o se especifican con `=true`/`=false`.
Las duraciones y rangos se validan antes de conectar a la base para ejecutar carga.

Ejemplo completo desde A hacia B, con `.env.b` que solo contiene las conexiones:

```bash
./seed.sh --env-file .env.b --manifest fixtures.b.json
./run.sh run --env-file .env.b --manifest fixtures.b.json --smoke --vus 1 --think-scale 0
./run.sh run --env-file .env.b --manifest fixtures.b.json --vus 110 \
  --summary-export b.summary.json --out json=b.results.json
# Al terminar la carga, con k6 detenido:
./cleanup.sh --env-file .env.b --run-id RUN_ID_DE_LA_EJECUCION
```

El launcher informa cuando compila y cuando inicia el comando. El seed muestra
la conexion, cada uno de los 10 fixtures, la validacion y la confirmacion de la
transaccion. Si no puede comunicarse con MySQL, devuelve un error y termina.
Los limites se configuran con flags, sin agregar variables al dotenv:

```bash
./seed.sh --env-file .env.b --manifest fixtures.b.json \
  --connect-timeout 10s --db-io-timeout 10s --operation-timeout 5m
```

Estos son los valores predeterminados. El limite de lectura/escritura nunca
supera el de conexion, para cubrir tambien el saludo y la autenticacion MySQL.
`--operation-timeout` limita seed/cleanup; no limita la duracion de la carga k6.
Ctrl+C cancela la conexion o el seed y revierte una transaccion sin confirmar.

Antes de la carga, k6 comprueba por HTTP que las diez ofertas precargadas esten
disponibles. Esta comprobacion no crea precondiciones ni consulta MySQL.
Los flags SQL (`--connect-timeout`, `--db-io-timeout`, `--operation-timeout`) solo
se admiten en los comandos independientes de mantenimiento. `run` no admite
`--preflight-timeout`: se elimino la validacion SQL previa.

El servidor externo debe tener Judge0 configurado y sus callbacks deben llegar
a su API. Se utiliza HTTPS para las cookies Secure. No se modifica el codigo,
el esquema ni los endpoints del servidor para ejecutar estas pruebas.

Los thresholds separan preflight de carga: errores HTTP <5%, errores funcionales
<1% globalmente y por accion; p95 <1.5s para registro/login/logout/busqueda/lectura,
participacion, pagina, enunciados, ejemplos, score, finalizacion y pulsaciones;
p95 de submission <8s. Se requiere al menos un recorrido completo y cobertura
de las acciones habilitadas. Se reportan p90/p95/p99.

Los VUs representan sesiones con pausas, no RPS constantes. Los percentiles
abarcan las rampas y el periodo estable; los puntos JSON permiten separarlos.
Una ejecucion abortada, sin cobertura o con recorridos interrumpidos no permite
concluir capacidad. No se ejecuta JavaScript de navegador ni se simula video.

## Servidor externo y verificaciones offline

`./seed.sh`, `./run.sh run` y `./cleanup.sh` no levantan servidores HTTP,
MySQL ni Judge0 locales. ./run.sh compila un ejecutable HTTP sin el codigo de
seed/cleanup ni el driver MySQL. El seed y la limpieza operan sobre la base configurada;
k6 envia la carga a BASE_URL y el servidor externo ejecuta los envios mediante
su Judge0. La carga y sus resultados corresponden a ese dispositivo.

Los auxiliares Python que levantaban HTTP simulado y MySQL temporal se retiraron.
Esas verificaciones comprobaban el runner; no medían la capacidad del servidor.

Opcionalmente, para comprobar el corpus y las reglas del runner sin servicios:

```bash
go test ./fixtures
node tests/offline.test.mjs
```

Los tests Go compilan las soluciones con g++ y javac (Java 13 o posterior),
verifican ejemplos/casos y errores de compilacion, y contrastan ambos lenguajes
con 60 entradas y oraculos independientes por lenguaje. Los compiladores son
necesarios solo para esta verificacion offline, no para lanzar la carga externa.
`k6 inspect` comprueba inicializacion y opciones sin ejecutar setup ni VUs.
Node verifica URL externa, manifiestos y veredictos SSE sin hacer peticiones.

Las comprobaciones SQL internas quedan excluidas de los tests habituales mediante
el build tag `capacity_internal_sql`; requieren un DSN explicito y tampoco
levantan MySQL. No forman parte de la medicion de capacidad. Ninguna verificacion
offline demuestra que el servidor externo soporte cierto numero de VUs: eso
requiere ejecutar `./run.sh run` contra su URL real.

## Verificacion directa en Judge0 existente

Esta comprobacion usa Judge0 real, sin la API de Spotted ni MySQL. No es una
medicion de capacidad. En este dispositivo se reutilizan los contenedores de
Judge0 1.13.1; necesita su API, workers, Redis y PostgreSQL. No iniciar los servicios
`app`, MySQL o mediamtx del compose de la raiz.

Desde `capacity_k6`, con los contenedores ya creados:

```bash
docker compose -f ../judge0/docker-compose-judge0.yml start db redis server workers
./verify-judge0.sh \
  --judge-url http://127.0.0.1:2358 \
  --judge-config ../judge0/judge0.conf \
  --report judge0.verification.json
```

`verify-judge0` no requiere DATABASE_URL ni carga `.env`. Lee autenticacion del
archivo indicado y la envia por headers, sin imprimir credenciales. Para una
API sin autenticacion, usar `--judge-config ''`.

Las entradas/salidas estan definidas en `fixtures/algorithms.go`, incluidas las
entradas grandes deterministas. Se envian directamente como `stdin` y
`expected_output` a `/submissions/batch`; no hay callbacks. Usa 1 segundo de CPU
y 262144 KB de memoria, los mismos limites almacenados por el seed. Consulta los
tokens hasta que el juez finaliza y comprueba cada caso individualmente:

- 90 casos secretos: tres algoritmos, dos lenguajes, tres variantes y cinco casos.
- 12 ejemplos: dos por algoritmo y lenguaje, solo para la variante correcta.
- Correctas: `Accepted` (3) y salida exacta, incluido salto de linea.
- Parciales: `Accepted` (3) o `Wrong Answer` (4) en los casos previstos.
- Sintaxis: `Compilation Error` (6), con diagnostico del compilador; no se acepta
  un error de ejecucion o de infraestructura como equivalente.

El comando falla si un resultado difiere y guarda el informe JSON con version,
limites, tokens, veredictos, salidas, diagnosticos, tiempos y memoria. En esta
instalacion `ENABLE_SUBMISSION_DELETE` esta deshabilitado: los envios quedan en
Judge0, y el verificador no cambia su configuracion para borrarlos. La limpieza
de las pruebas de capacidad sigue actuando sobre los datos temporales de Spotted.
