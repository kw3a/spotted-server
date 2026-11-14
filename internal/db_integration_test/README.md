# Integración servidor–MySQL

Estas pruebas ejecutan los métodos reales de `MysqlStorage` contra MySQL 8.0.
Siguen la estructura de los paquetes terminados en `_test`: funciones `Test…`,
subcasos con `t.Run` y aserciones de `testify/require`.

- `inc1`: submissions, callbacks y puntajes, participación y finalización,
  transacción de quiz, problemas completos y recuperación del último código.
- `inc2`: transacción de registro de ofertas, revisión de postulaciones,
  consultas por propietario y archivado.
- `testutil`: fixtures y preparación de bases vacías por caso.

Desde la raíz del repositorio, con Docker, Go y `goose` disponibles:

```bash
./internal/db_integration_test/run.sh -v
```

El script crea un contenedor nuevo, publica MySQL en un puerto aleatorio de
loopback y genera una contraseña temporal. No lee `.env`, no monta directorios
del proyecto y no reutiliza contenedores ni volúmenes existentes. Al salir,
incluso si falla una prueba, elimina su contenedor y sus volúmenes anónimos.

Cada caso crea su propia base, aplica las migraciones con Goose (primero schema,
luego triggers), prepara datos mínimos y elimina la base al terminar. No se
utiliza el seeder de desarrollo. Los casos de rollback tardío usan datos que
exceden los límites de columnas bajo el modo estricto de MySQL.

Se pueden filtrar pruebas mediante los argumentos habituales de Go:

```bash
./internal/db_integration_test/run.sh -v -run TestRegisterOffer
./internal/db_integration_test/run.sh -v -run TestUpdateTestCaseResult
```

Para un MySQL de pruebas administrado externamente, se admite
`DB_INTEGRATION_DSN` sin nombre de base y con permisos de creación y eliminación:

```bash
DB_INTEGRATION_DSN='user:password@tcp(127.0.0.1:3306)/?parseTime=true' \
  go test -count=1 ./internal/db_integration_test/inc1 ./internal/db_integration_test/inc2
```

Sin esa variable, `go test ./...` omite explícitamente estas pruebas; con una
variable configurada, una conexión o migración fallida hace fallar la suite.
`-count=1` evita reutilizar resultados de la caché frente a cambios externos.

`TestUpdateTestCaseResultRepeatedAccepted` comprueba que repetir un callback no
infle el puntaje. Se mantiene como regresión activa: un fallo señala el
comportamiento del trigger, no se omite ni se cambia su expectativa para aprobar.

Las pruebas de refresh tokens son transversales a ambos incrementos y no forman
parte de estas dos suites funcionales.

## Integración servidor–Judge0 (inc1)

```bash
./internal/db_integration_test/run-judge0.sh -v
```

Requiere estas imágenes **ya instaladas**: `judge0/judge0:1.13.1`,
`postgres:16.2` y `redis:7.2.4`. El ejecutor verifica su existencia antes de crear
recursos y usa `pull_policy: never` y `--pull never --no-build`. No descarga
imágenes, no modifica ni reutiliza contenedores existentes y no utiliza las
credenciales de `judge0.conf` o `.env` del proyecto. Crea configuración temporal,
una red, un volumen PostgreSQL y cuatro contenedores independientes. Los elimina
al finalizar, también cuando falla una prueba.

Las pruebas ejecutan `codejudge.Judge0.Send` contra Judge0 real. Los workers
llaman por HTTP al handler real `CreateCallbackHandler`, con router Chi,
decodificador JSON, decodificación base64 y `codejudge.Stream` reales. El
almacenamiento del callback se sustituye por un receptor en memoria para observar
el resultado sin introducir MySQL en esta integración.

Se comprueban:

- Aceptación de programas C++ y Java con entrada y salida conocidas.
- Respuesta equivocada, error de compilación, error de ejecución y tiempo límite.
- Un lote de dos casos con veredictos y salidas diferentes, relacionando los
  tokens con sus respectivos IDs de submission y caso.
- Finalización del stream y sus contadores después de los callbacks.
- Rechazo de credenciales inválidas por Judge0.

Cada caso tiene un plazo de 90 segundos y las peticiones salientes uno de 15.
La API se publica en un puerto aleatorio de loopback. El receptor de callbacks
escucha en la dirección del gateway de la red Docker temporal para que los
workers puedan alcanzarlo; el ejecutor está pensado para Docker local en Linux.
Judge0 utiliza `privileged`, como el compose existente del proyecto.

Para ejecutar contra servicios preparados externamente:

```bash
JUDGE0_INTEGRATION_URL='http://127.0.0.1:2358' \
JUDGE0_INTEGRATION_CALLBACK_HOST='172.18.0.1' \
JUDGE0_INTEGRATION_TOKEN='test-token' \
  go test -count=1 -timeout=10m ./internal/db_integration_test/inc1 -run '^TestJudge0' -v
```

La dirección de callbacks debe existir en el equipo donde corre Go y ser
alcanzable desde los workers. Sin `JUDGE0_INTEGRATION_URL`, estas pruebas se
omiten explícitamente. Una vez configurada, errores de comunicación o ejecución
hacen fallar las pruebas. No requieren `DB_INTEGRATION_DSN`.

## Integración aplicación web–MediaMTX (inc3)

```bash
./internal/db_integration_test/run-mediamtx.sh -v
```

Utiliza el binario local
`mediamtx_v1.17.0_linux_amd64/mediamtx` del repositorio y verifica que su versión
sea `v1.17.0`. No utiliza Docker ni descarga MediaMTX o navegadores. Se puede
indicar otra ubicación del mismo binario mediante `MEDIAMTX_INTEGRATION_BINARY`.
Requiere Node, Chrome y Playwright ya instalados. Admite `CHROME_BINARY` y
`PLAYWRIGHT_MODULE_PATH`; también busca Playwright en las dependencias locales o
en los drivers instalados de playwright-go (`~/.cache/ms-playwright-go`).

Cada prueba Go ejecuta el harness `inc3/browser.cjs`, que inicia una instancia
nativa independiente de MediaMTX, con puertos de loopback y un directorio temporal
para configuración y grabaciones. Usa segmentos de dos segundos. Al finalizar,
incluso ante una aserción fallida, cierra Chrome, apaga esa instancia y elimina sus
archivos temporales. No usa la configuración, los puertos ni las grabaciones de
la instalación existente.

El navegador carga los archivos reales `quiz-recording.js` y
`applicant-playback.js` desde una página mínima con los elementos DOM que
necesitan. Esa página no implementa publicación ni reproducción alternativas.
La cámara y el micrófono son los dispositivos sintéticos de Chrome; la captura
de pantalla se sustituye por un canvas coloreado. La composición del video,
WHIP, WebRTC, recepción de medios, grabación, descarga y reproducción son reales.
Estas pruebas no cubren la construcción de las páginas Go, autenticación o MySQL.

- `TestMediaMTXPublish`: conexión WebRTC, respuesta WHIP, pistas H264/Opus y bytes
  recibidos que aumentan en MediaMTX.
- `TestMediaMTXStopPublishing`: DELETE de WHIP exitoso, conexión cerrada,
  dispositivos liberados y publicación retirada de MediaMTX.
- `TestMediaMTXRecordAndPlayback`: publica al menos cinco segundos de medios,
  detiene, espera grabaciones disponibles y carga el reproductor real. Comprueba
  descarga por el ID correcto, video decodificado, tiempo que avanza y color.
- `TestMediaMTXParticipationIsolation`: graba una participación roja y otra azul;
  después de tener ambas grabaciones, reproduce cada ID y verifica su contenido.
- `TestMediaMTXUnavailable`: apaga la instancia antes de publicar y comprueba el
  mensaje de error, el botón de reintento, la conexión cerrada y pistas detenidas.
- `TestMediaMTXNoRecordings`: consulta una participación sin video y verifica el
  estado vacío, sin fuente de video ni segmentos.

Los casos se omiten en `go test ./...` si no se configura
`MEDIAMTX_INTEGRATION_BINARY`. Con el ejecutor se habilitan y fallan si falta el
binario, el navegador o las dependencias. Cada caso tiene un límite de dos
minutos; las esperas de condiciones son acotadas y conservan diagnósticos de
MediaMTX y del navegador al fallar.
