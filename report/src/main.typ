#import "template.typ": *
#import "utils.typ": *

#import "@preview/bytefield:0.0.1": *
#import "@preview/big-todo:0.2.0": *

#show: project.with(
  title: [Comunicações por Computador\ Grupo 69],
  authors: (
    (name: "Daniel Pereira", number: "A100545", photo: asset_path("danny.png")),
    (name: "Francisco Ferreira", number: "A100660", asset_path("francisco.png")),
    (name: "Rui Lopes", number: "A100643", photo: asset_path("rui.png")),
  ),
  logo: asset_path("logo.jpg"),
  date: "17 de dezembro de 2023",
)

#heading(numbering:none)[Introdução]
Este relatório tem como objetivo apresentar o trabalho prático desenvolvido para a unidade curricular de Comunicações por Computador. O trabalho consiste no desenvolvimento de um serviço de partilha de ficheiros _peer to peer_, não totalmente descentralizado. O relatório terá como objetivo apresentar a arquitetura do sistema, a sua descrição, implementação e as decisões tomadas durante o desenvolvimento do mesmo. 

Uma das liberdades cedida pela equipa docente foi a da escolha da linguagem de programação para o desenvolvimento do projeto. O nosso grupo decidiu optar por Golang #footnote[Arrependimento é real.].

= Arquitetura
Desde o começo da realização do projeto, tivemos a ambição de realizar todas as funcionalidades pedidas. Então, como pedido, o sistema conta com um _tracker_, responsável por controlar o estado atual da rede, _nodes_, que comunicam entre si para a realização das transferências e um sistema de servidores DNS. Cada uma destas vertentes irá ser detalhada, na respetiva secção, mais em baixo. Assim, apresentamos a seguir a arquitetura geral a que chegamos:

#figure(
  image(asset_path("arquitetura.png"), width: 75%),
  caption: [
    Arquitetura geral
  ],
)

Para organização e separação de responsabilidades, o projeto foi desenvolvido tendo por base o que são os _standards_ de Go. Assim, dentro da diretoria `cmd` encontra-se a implementação dos programas `tracker` e `node`. De outra forma, dentro da diretoria `internal` encontram-se todos os módulos auxiliares (e, por vezes, comuns) ao bom funcionamento de cada um destes programas. A execução de cada um destes programas e a descrição exaustiva do sistema de _building_ e configuração do mesmo encontram-se em espalhados por secções mais em baixo.

= Tracker
#let start_tracker_link = "https://github.com/ruilopesm/PessiTorrent-CC/blob/main/cmd/tracker/tracker.go#L30"
#let enqueue_link = "https://github.com/ruilopesm/PessiTorrent-CC/blob/main/internal/transport/tcp.go#L69"
#let tcp_handler = "https://github.com/ruilopesm/PessiTorrent-CC/blob/main/cmd/tracker/handlers.go#L10"
#let sync_map_link = "https://github.com/ruilopesm/PessiTorrent-CC/blob/main/internal/structures/sync_map.go#L5"
#let start_node_link = "https://github.com/ruilopesm/PessiTorrent-CC/blob/main/cmd/node/node.go#L53"

O _tracker_ é a peça essencial para o funcionamento deste serviço de partilha _peer to peer_. É ele o responsável por guardar o estado atual da rede. 

== Informação de estado
Assim, o mesmo possui duas estruturas de dados do tipo #link(sync_map_link)[`SynchronizedMap`] #footnote[_Wrapper_ genérico de `map` de Go com _mutexes_ embutidos, para o devido controlo de concorrência.], uma para guardar os ficheiros partilhados na rede e outra para guardar os dados referentes a cada _node_ presente na rede. A primeira é um mapeamento de nome de ficheiro para informação de ficheiro, onde estão presentes o tamanho de um ficheiro, a _hash_ de um ficheiro e a _hash_ de cada uma das _chunks_ de um ficheiro. A segunda é um mapeamento de endereço de _node_ para informação de _node_, onde estão presentes uma referência para a conexão mantida com esse _node_, a porta onde esse _node_ instanciou um servidor UDP e mais um `SynchronizedMap` que mapeia de nome de ficheiro para uma lista com os _chunks_ que esse _node_ possui do ficheiro.

== _Setup_ inicial
Ao iniciá-lo, é invocado o método #link(start_tracker_link)[```go Start()```], responsável por instanciar um servidor TCP, por via da função `net.Listen(...)` de Go. O método cria, ainda, uma _goroutine_ #footnote[Uma _goroutine_ pode ser descrita como uma _lightweight thread_ gerida pelo _runtime_ de Go.] responsável pela aceitação de conexões. Esta, é responsável por atender _nodes_ que se queiram conectar ao _tracker_ e criar um _wrapper_ da conexão estabelecida, um dedicado a cada _node_ conectado. Assim, é atingida concorrência máxima em interações futuras entre vários _nodes_ e o _tracker_. De notar que quando um _node_ sai da rede, ou seja, quando a conexão é fechada pelo _node_, toda a informação relacionada ao mesmo é eliminada. O _wrapper_ da conexão, quando iniciado, cria duas _goroutines_: uma de escrita e outra de leitura.

A _goroutine_ de escrita fica a ler de um _channel_ de Go, bloqueando sempre que o _channel_ não possui pacotes por ler. Ao ser colocado lá um pacote, através do método #link(enqueue_link)[```go EnqueuePacket(packet protocol.Packet)```], essa _goroutine_ consome-o, serializa-o para a conexão e faz _flush_.

A _goroutine_ de leitura fica à espera de receber um pacote no _socket_, _socket_ este instanciado pelo servidor TCP. Após a leitura de um dado pacote, é invocado o _handler_ de pacotes, recolhido pelo construtor do servidor TCP. Este _handler_ deverá ser uma função respeitadora do seguinte tipo `type TCPPacketHandler func(packet protocol.Packet, conn *TCPConnection)`. Esta arquitetura permite que seja muito simples criar e utilizar diferentes _handlers_. Vale também relembrar que cada pacote é _handled_ numa _goroutine_ separada, permitindo, assim, que o _tracker_ consiga responder a vários pedidos de forma concorrente. Neste caso em específico, o _handler_ utilizado pode ser encontrado #link(tcp_handler)[aqui].

== Execução
Decidimos que o _tracker_ não precisava de uma CLI #footnote(link("https://en.wikipedia.org/wiki/Command-line_interface")) como base. A execução do mesmo é tão simples quanto a execução do comando `./out/bin/tracker` #footnote[A geração deste executável é responsabilidade da _Makefile_, através do comando `make tracker`.]. Este comando conta com uma _flag_ opcional: `-p`, onde pode ser passada uma porta específica para a criação do servidor TCP #footnote[A porta por defeito é a 42069.].

= Node
#let enqueue_request_link = "https://github.com/ruilopesm/PessiTorrent-CC/blob/main/internal/transport/udp.go#L106"

O _node_ é simplesmente uma abstração para um participante da rede. São estes os responsáveis por alimentar a rede, no sentido de poderem publicar nela ficheiros. Além disso, podem também solicitar ficheiros a outros _nodes_.

== Execução
A execução de um _node_ é dada pelo comando `./out/bin/node` #footnote[A geração deste executável é responsabilidade da _Makefile_, através do comando `make node`.], que conta com duas _flags_ opcionais: `-t`, onde pode ser passado o endereço e porta onde se encontra o servidor TCP do _tracker_ e `-p`, onde pode ser passada uma porta específica para a criação do servidor UDP #footnote[A porta por defeito é a 8081.].

== _Setup_ inicial
Quando o programa _node_ inicia, é invocado o método #link(start_node_link)[```go Start()```], tal como no caso do _tracker_. Este método é responsável por criar quatro _goroutines_ com papéis importantes para o bom desempenho de um _node_.

A primeira _goroutine_ é responsável por contactar, por meio da função `net.Dial(...)` de Go, o _tracker_ e estabelecer uma conexão com o mesmo. Da mesma forma que no _tracker_, é novamente criado um _wrapper_ para esta conexão com duas _goroutines_. No entanto, existe a remota possibilidade do _tracker_ não estar disponível e, portanto, um cuidado que tivemos foi que o programa _node_ não depende totalmente da existência do _tracker_. Assim, alguns comandos (que iremos detalhar mais em baixo) estão disponíveis mesmo quando o _tracker_ não existe na rede.

A segunda é responsável por criar um servidor UDP, utilizado pelo _node_ para enviar e receber pedidos relacionados a transferências. Quando iniciado, o servidor conta, tal como no servidor TCP do _tracker_, com duas _goroutines_ principais: escrita e leitura. \
A _goroutine_ de escrita fica a ler de um _channel_ de Go, totalmente voltado para _requests_ de _chunks_, novamente, bloqueando sempre que o _channel_ não possui _requests_ por ler. Cada _request_ é composto por um pacote (`protocol.Packet`) e um endereço UDP para onde é suposto enviar o _request_. Após um pacote ser colocado no dito _channel_, através do método #link(enqueue_request_link)[```go EnqueueRequest(packet protocol.Packet, addr *net.UDPAddr)```], a _goroutine_ de escrita consome-o, serializa-o e, finalmente, escreve-o pelo _socket_ UDP. \
Por outro lado, a _goroutine_ de leitura comporta-se, também, da exata mesma forma que no caso do servidor TCP do _tracker_, mas desta vez o _handler_ utilizado deve respeitar o tipo `type UDPPacketHandler func(packet protocol.Packet, addr *net.UDPAddr)`. No nosso caso, a definição do mesmo pode ser encontrada #link("https://github.com/ruilopesm/PessiTorrent-CC/blob/main/cmd/node/handlers.go#L30")[aqui].

#pagebreak()
== CLI
Uma vez que faz todo o sentido que este programa seja suportado por uma CLI, a terceira _goroutine_ é responsável pela criação e execução da mesma. Esta CLI conta com uma API onde é possível registar comandos, com nome, _usage_, descrição, número de argumentos e a respetiva função a executar. Os comandos são lidos a partir do _input_ do utilizador, tal como demonstrado imediatamente em baixo.

```sh
Type 'help' for a list of available commands
UDP server started on 0.0.0.0:8083
Connected to tracker on 127.0.0.1:42069
> help
Available commands:
  - connect	<tracker address>	Connect to the tracker
  - publish	<file name>	
  - request	<file name>
  - remove  <file name>
  - status	Show the status of the node
  - statistics	Show the statistics of the node
  - help	Show this help
  - exit	Exit the program
```

Como existirão mensagens a serem escritas enquanto o utilizador está a escrever o seu _input_ (e.g. receber informações relativas à transferência de um ficheiro anteriormente pedido), foi necessário recorrer à biblioteca `x/term` #footnote(link("https://pkg.go.dev/golang.org/x/term@v0.15.0")) para resolver o problema do _stdout_ escrever à frente do _input_ do utilizador.

De seguida, irão ser detalhados cada um dos comandos e a sua respetiva função.

- `connect <tracker address>` - estabelecer a conexão com o tracker, caso o mesmo não exista no momento de instanciação do _node_ (tal como descrito na anteriormente).
- `publish <file name | directory>` - publicar um ficheiro ou diretoria (de forma recursiva) na rede.
- `request <file name>` - pedir um ficheiro à rede.
- `remove <file name>` - remover um ficheiro anteriormente publicado.
- `status` - mostrar alguns dados relevantes ao nodo, como o seu estado de conexão ao _tracker_ ou a percentagem a que se encontram as transferências a decorrer.
- `statistics` - mostrar alguns dados relevantes às transferências, como o total de _bytes_ _uploaded_ e _downloaded_, bem como a velocidade média de transferência.
- `help` - tal como o nome indica, mostrar a listagem de comandos.
- `exit` - sair do programa e consequentemente da rede.

== Informação de estado
#let tick_link = "https://github.com/ruilopesm/PessiTorrent-CC/blob/main/cmd/node/node.go#L143"

O mesmo conta com quatro estruturas de dados do tipo `SynchronizedMap`. Uma para os ficheiros cujo estado é _pending_, uma para os ficheiros cujo estado é _published_, outra para os ficheiros cujo estado é _downloading_ e, finalmente, outra para os ficheiros cujo estado é _downloaded_. No _map_ de ficheiros _downloading_, em cada _value_ é possível encontrar uma estrutura que, para além dos dados habituais de um ficheiro, possui também a última vez que foi solicitado um _update_ ao _tracker_ sobre esse ficheiro (como forma de saber possíveis novos _nodes_ que tenham _chunks_ desse ficheiro). Cada _node_ possui ainda informação relativa a todos os _nodes_ relevantes, para si, da rede, sobre os _chunks_ que possuem de um determinado ficheiro cujo estado seja _downloading_ e o número de _timeouts_ dados durante a comunicação (tentativa de comunicação) com o mesmo. Cada _chunk_ possui informação sobre a última vez que foi pedido e o número de tentativas, valor que é afetado quando o _node_, a quem foi pedido, dá _timeout_ ou o pacote possui erros. \
Para além destas quatro referidas estruturas, o _node_ conta ainda com uma estrutura de estatísticas onde são guardados dados relevantes para o cálculo do tempo médio de transferência entre dois _nodes_. Obviamente, esta informação não poderia ser guardada pelo _tracker_, visto que este tempo é relativo. \
Toda esta informação é utilizada, principalmente, pelo sistema de _ticks_ (detalhado na secção imediatamente em baixo) e, consequentemente, pelo algoritmo de escalonamento (detalhado numa secção mais em baixo) na escolha de quais _chunks_ e a quais _nodes_ pedir primeiro. \

== Sistema de ticks
Por fim, a última _goroutine_ é totalmente dedicada a um sistema de _ticks_. Este sistema executa uma determinada tarefa a cada _tick_ do programa, definido para (valor configurável) 100 milissegundos. Essa tarefa deve ser pura, no sentido de que para um dado _input_,  o seu _output_ deverá ser sempre o mesmo (algo que acontece). Atualmente, a tarefa/função executada a cada tick _roda_ o algoritmo de escalonamento e trata de todo o sistema de _timeouts_. Foi bastante mais fácil e proveitoso desenvolver um sistema de _timeouts_ baseado em _ticks_, isto pois, as duas alternativas em que pensamos seriam: _polling_ constante à informação de estado do _node_ ou todo um sistema complexo de comunicação entre uma _goroutine_ principal e outras _goroutines_ (e respetivos _sockets_) cada uma com o seu _timeout_. Em relação à primeira alternativa é facilmente percetível que coloca bastante carga desnecessária sobre o sistema e sobre estruturas que utilizam mecanismos de exclusão mútua (_mutexes_), o que tornaria todo o processo de transferência mais lento. A segunda alternativa, apesar de minimamente interessante, traz uma dificuldade e confusão excessiva #footnote[Podendo ser apelidada de solução _finished_.] a todo o programa.

= Algoritmo de escalonamento
Dentro do sistema de _ticks_ descrito acima, o algoritmo de escalonamento é chamado para execução. Para cada ficheiro, organiza-se as chunks pela raridade delas. A raridade é determinada pelo número de nodos que têm essa chunk disponível para download. Então, de forma a cumprir boas práticas de sistemas distribuúdos, as chunks com mais raridade serão descarregadas com mais prioridade. Após determinado as chunks mais relevantes, pedimos $M$ chunks a cada nodo, sem haver repetidos, por ordem de velocidade de transferência média. Este $M$ foi estipulado para 100 chunks. Ou seja, a cada iteração de tick, o algoritmo pedirá no máximo 100 chunks a cada nodo em que as mais relevantes serão pedidas aos nodos com maior velocidade.

Esses 100 chunks são agrupados num só #link(<request_chunk_packet>)[pacote de pedido] para cada nodo.

A velocidade de transferência média para cada nodo é calculada a partir da média de $"chunkSize"/"latência"$ de todos os pacotes recebidos desse nodo nos últimos 100 segundos (valor estipulado).

Caso aconteça outra chamada ao tick em que há chunks pendentes que ainda não chegaram passado 500 milissegundos (valor estipulado), é reenviado o pedido e é incrementado uma penalização na chunk. Ao fim de 3 (valor estipulado) timeouts numa mesma chunk, é incrementado uma penalização no nodo, e, ao fim de 3 (valor estipulado) dessas penalizações o nodo é removido da lista de nodos que têm o ficheiro, podendo só voltar ao receber uma atualização de informações do ficheiro do _tracker_ (assume-se que o _tracker_ não irá enviar nodos offline).

O algoritmo de escalonamento foi a parte do trabalho onde achámos que existem muitas melhorias por fazer, como falado na @limitations.

= Ficheiros
Num sistema de partilha de ficheiros _peer to peer_ era de esperar que existisse bastante tempo dedicado à forma como os ficheiros são geridos. \
Uma das nossas maiores preocupações, desde início, foi garantir a integridade do conteúdo de um ficheiro que seja transferido. Uma vez que o protocolo de transferência (detalhado mais em baixo) funciona sobre UDP, não existe garantia de _error checking_ e/ou _error correction_. Para tal, foi necessário, ao nível aplicacional, implementar um mecanismo de _hashing_. \
Além disso, outra preocupação foi também a divisão de um ficheiro em _chunks_, isto pois transferir ficheiros de grande dimensão num só chunk é impensável num ambiente em que perdas de pacotes são constantes e em que existe um tamanho máximo (neste caso, imposto pelo UDP).

== Hashing
Para o _hashing_ decidimos utilizar o algoritmo SHA-1. Esta decisão centra-se na probabilidade muito reduzida de colisão #footnote[$(1 / 2 * n^2) / 2^160$, onde $n$ é o número de _hashes_ geradas.] e no tamanho, não muito grande, da _hash_ gerada - 20 _bytes_. \
No nosso caso, cada _chunk_ tem a sua própria _hash_. Estas _hashes_ são calculadas quando um _node_ quer publicar um ficheiro e enviadas para o _tracker_. Assim sendo, é este que detem todas as _hashes_ presentes na rede, fazendo com que não seja possível alguns participantes enganarem outros.

== Chunks
#let kB(bits) = bits * 8000
#let mB(bits) = kB(bits * 1000)
#let gB(bits) = mB(bits * 1000)

A decisão de qual o número de _bits_ utilizar para codificar um _chunk_ prende-se em encontrar um bom _sweet spot_ entre o tamanho de um ficheiro e o número de chunks. Não queremos que um ficheiro pequeno tenha imensos _chunks_ - o que causaria _overhead_ - nem que um ficheiro grande tenha poucos _chunks_ resultando em _chunks_ maiores, o que causaria muito transtorno aquando da perda de um pacote ou aquando de um pacote com erros. Para isso, o tamanho de um _chunk_ não pode ser estático. Assim, desenvolvemos a tabela em baixo presente para estudarmos um pouco o caso.

#let file_sizes = (
  (kB(100), "100 kB"),
  (kB(500), "500 kB"),
  (mB(1), "1 MB"),
  (mB(10), "10 MB"),
  (mB(100), "100 MB"),
  (gB(1), "1 GB"),
  (gB(10), "10 GB"),
  (gB(100), "100 GB"),
)

#let chunk_sizes = (
  (kB(16), "16 kB"),
  (kB(64), "64 kB"),
  (kB(256), "256 kB"),
  (kB(512), "512 kB"),
  (mB(1), "1 MB"),
  (mB(2), "2 MB"),
  (mB(4), "4 MB"),
  (mB(8), "8 MB"),
  (mB(16), "16 MB"),
)

#figure(
  tablex(
    align: center + horizon,
    columns: file_sizes.len() + 1,
    cellx(fill: gray.lighten(40%), bdiagbox[Chunk][File]), 
    ..file_sizes.map(v => cellx(fill: gray.lighten(60%), v.at(1))),
    ..chunk_sizes.map(chunk => {
      (cellx(fill: gray.lighten(60%), chunk.at(1)),) + file_sizes.map(file_size => cellx(calc.ceil(file_size.at(0) / chunk.at(0))))
    }).flatten()
  ), 
  kind: table
)

Esta tabela apresenta a relação entre um dado tamanho de _chunk_ e um dado tamanho de ficheiro. Por exemplo, para um tamanho de _chunk_ de 16 kB e um tamanho de ficheiro de 100 MB seriam necessários, aproximadamente, 6250 _chunks_. Agora, só precisamos de saber quantos _bits_ precisamos para identificar cada quantidade de _chunks_. Para isso, desenvolvemos outra tabela, apresentada em baixo.

#figure(
  tablex(
    align: center + horizon,
    columns: file_sizes.len() + 1,
    cellx(fill: gray.lighten(40%), bdiagbox[Chunk][File]), 
    ..file_sizes.map(v => cellx(fill: gray.lighten(60%), v.at(1))),
    ..chunk_sizes.map(chunk => {
      (cellx(fill: gray.lighten(60%), chunk.at(1)),) + file_sizes.map(
        file_size => cellx(
          calc.ceil(calc.log(calc.ceil(file_size.at(0) / chunk.at(0)), base: 2))
        )
      )
    }).flatten()
  ), 
  kind: table
)

Desta vez, a tabela apresenta o número de _bits_ necessários para representar um determinado número de _chunks_ (valor derivado da tabela anterior). Por exemplo, para um tamanho de _chunk_ de 256 kB e um tamanho de ficheiro de 1 GB são necessários 12 _bits_ para representar na totalidade os diferentes _chunks_, isto pois, $2^12 (4096) > 3907$, mas $2^11 (2048) < 3907$.

#let chunk_id_bit_size = 16.

Observando ambas as tabelas, decidimos que *#chunk_id_bit_size _bits_* seria uma boa escolha. Tomamos, ainda, a liberdade de desenvolver mais uma tabela, desta vez para percebermos qual o tamanho máximo de ficheiro para um dado _bit size_, neste caso #chunk_id_bit_size _bits_.

#figure(
  tablex(
    align: center + horizon,
    columns: chunk_sizes.len(),
    ..chunk_sizes.map(chunk => cellx(fill: gray.lighten(60%), chunk.at(1))),
    ..chunk_sizes.map(chunk => {
      text(size: 9pt)[#calc.round(digits: 2, chunk.at(0) * calc.pow(2, chunk_id_bit_size)/gB(1)) GB]
    })
  ), 
  kind: table
)

Para o envio de chunks entre nodos, como visto acima, o tamanho de cada chunk deverá ser variável para acomodar vários tamanhos de ficheiro. Assim, durante as comunicações, em vez de ser enviado o tamanho completo (para um ficheiro de 16 kB, teríamos que enviar o valor 128000, o tamanho de 16 kB em bits), podemos apenas enviar $c/(16 op("kB"))$, com $c$ sendo o tamanho da chunk. E, portanto, é possível derivar, facilmente, o tamanho do ficheiro a partir do tamanho da _chunk_ e vice-versa, através da fórmula apresentada de seguida:

$ c = ceil(t/(2^b*16 op("kB"))) * 16 op("kB") $

$c$ - tamanho de cada chunk

$t$ - tamanho do ficheiro

$b$ - número de bits para identificação de cada chunk (no nosso caso 16)

$ceil(n)$ - número $n$ arredondado para cima

== Persistência de chunks em disco
#let enqueue_chunk_link = ""
Aquando da receção de um _chunk_ pedido, é necessário persistir o mesmo de alguma forma no _node_. Uma das hipóteses seria acumular os _chunks_ recebidos numa estrutura em memória e após a transferência de um dado ficheiro escrevê-los todos em disco. No entanto, esta abordagem tem dois problemas: ficheiros grandes iriam ocupar (ou até mesmo nem caber) bastante memória RAM; não existiria nenhum nível de concorrência na escrita para disco. \
Assim, decidimos criar uma alternativa que resolve ambos os problemas. Para isso, criamos um módulo coordenado por uma estrutura #link("https://github.com/ruilopesm/PessiTorrent-CC/blob/main/internal/filewriter/filewriter.go#L18")[`FileWriter`] que, quando iniciada, cria uma _pool_ de 10 _goroutines_ prontas a consumir de um _channel_ de _chunks_ a escrever. Este _channel_ é alimentado pelo método #link(enqueue_chunk_link)[```go EnqueueChunkToWrite(index uint16, data []byte)```]. Assim, é possível escrever, em disco, 10 _chunks_ ao mesmo tempo sem nenhum problema, isto pois, factualmente os _chunks_ a escrever começam e acabam sempre em posições diferentes do ficheiro. Aproveitamos também e fizemos com que este `FileWriter` mantenha um _file descriptor_ durante toda a transferência de um dado ficheiro - o que faz com que não tenhamos de abrir o ficheiro a cada escrita que fazemos. Para tal, quando um `FileWriter` é instanciado, é criado um _sparse file_ #footnote(link("https://en.wikipedia.org/wiki/Sparse_file")) e guardado o respetivo descritor de ficheiro durante o tempo de vida do `FileWriter`, como referido anteriormente.

= Serialização e Desserialização
#let serialize_link = "https://github.com/ruilopesm/PessiTorrent-CC/blob/main/internal/protocol/serialization.go#L10"
#let desserialize_link = "https://github.com/ruilopesm/PessiTorrent-CC/blob/main/internal/protocol/serialization.go#L62"

Antes de avançarmos para a descrição de cada um dos protocolos implementados durante o projeto (talvez a parte mais importante), faz todo o sentido abordarmos a serialização e desserialização de todos os pacotes a serem enviados. \ 
O método de serialização desenvolvido no nosso projeto é comum a ambos o _node_ e o _tracker_. Para este efeito, recorremos à utilização de _reflection_#footnote[Capacidade de um programa examinar a sua própria estrutura.], o que nos permitiu criar um módulo que facilmente realiza a serialização e desserialização dos nossos pacotes, tornando assim muito mais fácil e eficaz o seu processo de criação e alteração. Por exemplo, basta criar um nova `struct` respeitadora da interface `protocol.Packet` para que a sua serialização e desserialização sejam automaticamente implementadas, algo bastante conveniente. \
A interface deste módulo consiste apenas nas funções #link(serialize_link)[`SerializePacket()`] e na #link(desserialize_link)[`DeserializePacket()`].

==  Bitfield
Como demonstrado anteriormente, por vezes, o número de _chunks_ pode ser muito alto. Imaginando um caso em que um _node_ pede todos os _chunks_ de um ficheiro de 500 MB a outro, é fácil de perceber que isso se traduz num pacote bastante pesado, composto, praticamente de números (os índices de cada _chunk_). Como tal, tivemos a ideia de codificar os _chunks_ que um dado _node_ dispõe de um dado ficheiro num bitfield. Um bitfield consiste num _array_ onde cada elemento apenas pode tomar os valores de um _bit_, 0 ou 1. No nosso caso em específico, quando um elemento está a 0 significa que o _node_ possui o _chunk_ desse índice, o contrário quando está a 1. Por exemplo, o bitfield `110111011110` indica-nos que o _node_ possui todos os _chunks_ do ficheiro exceto os _chunks_ 3, 7 e 12. Este mecanismo ajudou-nos imenso a poupar memória e tornar todas as interações mais rápidas.

= Formato das mensagens protocolares <7>

As mensagens entre nodos e _tracker_ são enviadas e lidas em formato binário em little-endian. Para simplificação na descrição dos pacotes é assumido a existência dos seguintes tipos:
- Inteiros:
  - u8: Inteiro unsigned com 8 bits
  - u16: Inteiro unsigned com 16 bits
  - u32: Inteiro unsigned com 32 bits
  - u64: Inteiro unsigned com 64 bits
- Tipos compostos:
  - [$T$]: Array dinàmica de $T$, sendo $T$ qualquer tipo
    - Enviado com um u32 a representar o tamanho da array, seguido dos $T$ da array serializados
  - [$T$; $S$]: Array tamanho fixo $S$ de tipo $T$, sendo $T$ qualquer tipo
    - Enviado $T$ da array serializados
  - String: String em formato UTF-8
    - Enviado com um u32 a representar o tamanho da string, seguido da string em formato UTF-8
  - Bitfield: Bitfield em formato de array
    - Byte array de [u8], onde cada bit de cada u8 representa o valor (true ou false) naquela posição
  
- Tipos de dados exclusivos do programa: 
  - NodeFileInfo com campos (por ordem):
    - Name: String
    - Port: u16
    - Bitfield: Bitfield 
    
= FS Track
FS Track é o protocolo utilizado para a comunicação entre o _tracker_ e os _nodes_. O mesmo está implementado em cima de TCP.

== Interações
Este protocolo inicia-se sempre que um _node_ envia um #link(<InitPacket>)[`InitPacket`] ao _tracker_, de forma a informá-lo que pretende participar na rede. Após isso, o _tracker_ guarda as informações do _node_, passando este a estar apto para realizar publicações e pedidos de ficheiros. De notar, que como estamos a atuar em cima do TCP, não foi necessário realizar nenhum tipo de _handshake_.

A partir deste ponto, o _node_ pode publicar um ficheiro através do #link(<PublishFilePacket>)[`PublishFilePacket`]. Ao receber este este pacote, o _tracker_ verifica se um ficheiro com o nome especificado já existe. Se existir, é enviado um #link(<AlreadyExistsPacket>)[`AlreadyExistsPacket`] como resposta ao _node_. Caso não exista um ficheiro com o mesmo nome, o _tracker_ armazena as informações do ficheiro, atualiza também a informação relativa aos _nodes_ que possuem _chunks_ desse ficheiro e envia um #link(<FileSuccessPacket>)[`FileSuccessPacket`] (com tipo `PublishFileType`) como resposta.

Estando o ficheiro disponível no _tracker_, outros _nodes_ da rede podem requisitar o seu _download_, através do #link(<RequestFilePacket>)[`RequestFilePacket`], especificando apenas o nome do ficheiro que pretendem. Ao receber este pedido, o _tracker_ procura por todos os _nodes_ que possuem o ficheiro, parcialmente ou totalmente, e envia como resposta um #link(<AnswerFileWithNodesPacket>)[`AnswerFileWithNodesPacket`], onde consta toda a informação relativa ao ficheiro e aos _nodes_ que possuem esse mesmo ficheiro. Caso o ficheiro não exista na rede, o _tracker_ responde simplesmente com um #link(<NotFoundPacket>)[`NotFoundPacket`].

Durante a transferência do ficheiro pedido, o _node_ atualiza periodicamente o _tracker_ referindo que já tem disponíveis para _download_ uns dados _chunks_. Para isso, ele envia um #link(<UpdateChunksPacket>)[`UpdateChunksPacket`]. Assim, da próxima vez que um _node_ solicitar informação de um ficheiro ao _tracker_, terá a informação mais recente possível. \
Além disso, existe também o pacote #link(<UpdateFilePacket>)[`UpdateFilePacket`], que é, também, utilizado durante a transferência de um certo ficheiro. Esse pacote serve para solicitar ao _tracker_ a informação mais recente desse dado ficheiro. Nesse caso, o _tracker_ responde com um #link(<AnswerNodesPacket>)[`AnswerNodesPacket`], que se diferencia do #link(<AnswerFileWithNodesPacket>)[`AnswerFileWithNodesPacket`] na medida em que o primeiro apenas tem informação sobre os _nodes_ que possuem um ficheiro.

Finalmente, um _node_ pode remover um ficheiro enviando um #link(<RemoveFilePacket>)[`RemoveFilePacket`] ao _tracker_. Novamente, caso o ficheiro não exista na rede, o _tracker_ responde com um #link(<NotFoundPacket>)[`NotFoundPacket`]. Caso exista, o _tracker_ responde com um #link(<FileSuccessPacket>)[`FileSuccessPacket`] (com tipo `RemoveFileType`).

Novamente pelo facto de estarmos a atuar em cima de TCP, não é necessário cuidado com perda de pacotes, duplicação de pacotes ou erros em pacotes. É a camada de transporte, TCP no caso, que trata disso.

#pagebreak()
== Especificação das mensagens protocolares
_Ver @7 para entender melhor a definição dos tipos de dados._

As tabelas a seguir contém informação do que cada pacote leva. Os campos estão ordenados.

<InitPacket>
#packet_def(0, "InitPacket", "Node", "Tracker",
  [Name], [String], [Nome pelo qual é conhecido o node],
  [UDPPort], [u16], [Porta UDP do node]
)

<PublishFilePacket>
#packet_def(1, "PublishFilePacket", "Node", "Tracker",
  [FileName], [String], [Nome do ficheiro a publicar],
  [FileSize], [u64], [Tamanho em bytes do ficheiro],
  [FileHash], [[u8;20]], [Hash do ficheiro completo],
  [ChunkHashes], [[[u8;20]]], [Array de hashes de todas as chunks]
)

<FileSuccessPacket>
#packet_def(2, "FileSuccessPacket", "Tracker", "Node",
  [FileName], [String], [Nome do ficheiro a publicar],
  [Type], [u8], [A que tipo de pedido se refere o pacote, podendo ser `PublishFileType` ou `RemoveFileType`]
)

<AlreadyExistsPacket>
#packet_def(3, "AlreadyExistsPacket", "Tracker", "Node",
  [FileName], [String], [Nome do ficheiro que já se encontra publicado na rede]
)

<NotFoundPacket>
#packet_def(4, "NotFoundPacket", "Tracker", "Node",
  [FileName], [String], [Nome do ficheiro que não foi encontrado na rede]
)

<UpdateChunksPacket>
#packet_def(5, "UpdateChunksPacket", "Node", "Tracker",
  [FileName], [String], [Nome do ficheiro que irá ser atualizado],
  [Bitfield], [Bitfield], [Bitfield contendo os chunks que o node tem do dito ficheiro]
)

<RequestFilePacket>
#packet_def(6, "RequestFilePacket", "Node", "Tracker",
  [FileName], [String], [Nome do ficheiro requisitado]
)

<UpdateFilePacket>
#packet_def(7, "UpdateFilePacket", "Node", "Tracker",
  [FileName], [String], [Nome do ficheiro sobre o qual o node quer receber novas informações],
)

<AnswerFileWithNodesPacket>
#packet_def(8, "AnswerFileWithNodesPacket", "Tracker", "Node",
  [FileName], [String], [Nome do ficheiro a que o pacote diz respeito],
  [FileSize], [u64], [Tamanho do ficheiro],
  [FileHash], [[u8;20]], [Hash do ficheiro],
  [ChunkHashes], [[[u8;20]]], [Array de hashes de todas as chunks],
  [Nodes], [[NodeFileInfo]], [Array de informação de nodesinformação de nodes]
)

<AnswerNodesPacket>
#packet_def(9, "AnswerNodesPacket", "Tracker", "Node",
  [FileName], [String], [Nome do ficheiro a que o pacote diz respeito],
  [Nodes], [[NodeFileInfo]], [Array de informação de nodes]
)

<RemoveFilePacket>
#packet_def(10, "RemoveFilePacket", "Node", "Tracker",
  [FileName], [String], [Ficheiro a ser removido]
)

= FS Transfer
FS Transfer é o protocolo utilizado para a comunicação entre _nodes_ da rede. O mesmo está implementado em cima de UDP. Deste modo, ao contrário do FS Track foi necessário lidar com perdas de pacotes, duplicação de pacotes ou erros em pacotes, já que estes mecanimos não se encontram implementados em UDP.

== Interações
Este protocolo foi pensado e desenhado de forma a ser o mais simples possível. Uma das nossas preocupações foi sempre que o RTT (Round-trip time) fosse o menor possível. Como tal, não existem _acknowledgements_ explícitos, que, outrora, contribuíriam para o aumento desta métrica. \
Sempre que um _node_ já tem na sua posse informação sobre quais _nodes_ possuem um certo ficheiro (informação esta solicitada ao _tracker_), o mesmo envia um `RequestChunksPacket` a cada um dos _nodes_ em questão (a ordem é ditada pelo algoritmo de escalonamento, já detalhado). Se tudo correr bem, os _nodes_ a quem o pedido foi feito irão enviar um `ChunkPacket` por cada _chunk_ em questão, onde consta o conteúdo do _chunk_. No entanto, várias coisas podem correr mal e é nesse sentido que o protocolo implementa alguns mecanismos, descritos já de seguida. \
Caso um _node_ peça um _chunk_ e o mesmo venha com erros, este é simplesmente discartado, com a segurança de que no próximo _tick_ voltará a ser pedido. A verificação de erros é simplesmente a comparação da _hash_ fornecida pelo _tracker_ com a _hash_ do conteúdo vindo no pacote. \
Caso um _node_ receba um pacote de um _chunk_ que já havia marcado como _downloaded_, o mesmo é, novamente, descartado. \
Por último, caso um _node_ peça um _chunk_, mas o mesmo não chegue em tempo útil é efetuada toda uma lógica por parte do _ticker_ para que os _chunks_ sejam pedidos novamente aos mesmos (ou novos) _nodes_. Essa lógica já foi, anteriormente, descrita no algoritmo de escalonamento.

== Descrição das mensagens protocolares

<request_chunk_packet>
#packet_def(11, "RequestChunksPacket", "Node", "Node",
  [FileName], [String], [O nome do ficheiro a que o pacote diz respeito],
  [Chunks], [[u16]], [Array de números de chunks a pedir]
)

#packet_def(12, "ChunkPacket", "Node", "Node",
  [FileName], [String], [O nome do ficheiro a que o pacote diz respeito],
  [Chunk], [u16], [O número do chunk em questão],
  [ChunkContent], [[u8]], [Conteúdo em bytes do chunk]
)

= Serviço de resolução de nomes
Tal como foi pedido, adicionamos um serviço de resolução de nomes ao nosso projeto. Com isso, conseguimos identificar nodos na rede através do seu nome, invés de termos de utilizar os seus IPv4's. Isto facilita a utilização do sistema como um todo para o utilizador final, mas também para os desenvolvedores, pois é mais simples ler _logs_ produzidos com nomes. \
Para tal, recorremos à tecnologia _bind9_, bastante flexível e customizável.

Decidimos configurar dois servidores DNS, um primário e um secundáro, de forma a termos uma maior disponibilidade no serviço de resolução de nomes assim como distribuição de carga entre ambos os servidores, levando a uma maior rapidez nas respostas às _queries_ feitas e uma maior tolerância a faltas. Assim, caso um dos servidores falhe temos o outro como _backup_.

A configuração do servidor DNS começa com o #link(<options>)[`ns1.named.conf.options`], onde constam as configurações gerais de ambos os servidores.
No nosso caso, apenas especificamos a diretoria para os ficheiros de cache do bind e o DNS Security Extensions (DNSSEC) em modo automático.

Para especificarmos as zonas disponíveis na nossa rede, tivemos de configurar os ficheiro #link(<ns1>)[`ns1.named.conf.local`] e #link(<ns2>)[`ns2.named.conf.local`], que são respetivamente do servidor primário e secundário e que referenciam o ficheiro #link(<local>)[`local.zone`], para os _DNS lookups_, e #link(<reverse>)[`reverse.zone`], para os _reverse DNS lookups_.\
No nosso programa, a funcionalidade de _reverse DNS lookups_ possibilita que um _node_ seja capaz de deduzir automaticamente o seu nome com base no seu endereço IP, eliminando a necessidade de o utilizador inserir manualmente o nome associado à máquina durante o comando de inicialização do programa. Desta forma, reduzimos a possibilidade de erros e aumentamos a experiência de utilização da nossa aplicação. 

= Testes e Resultados

Mostraremos agora os resultados finais:

Outputs em cada programa:
#figure(caption: "Tracker 1", supplement: "Entidade")[  
```bash
$ ./out/bin/tracker
- TCP server started on 0.0.0.0:42069
- Node 127.0.0.1:40528 connected
- Init packet received from 127.0.0.1:40528
- Registered node with data: [127 0 0 1], 8081
- Publish file packet received from 127.0.0.1:40528
- Node 127.0.0.1:58746 connected
- Init packet received from 127.0.0.1:58746
- Registered node with data: [127 0 0 1], 8082
- Request file packet received from 127.0.0.1:58746
- Publish chunk packet received from 127.0.0.1:58746
```
]

#figure(caption: "Nodo que faz download", supplement: "Entidade")[  
```bash
$ ./out/bin/node -p 8082
- Type 'help' for a list of available commands
- UDP server started on 0.0.0.0:8082
- Connected to tracker on 127.0.0.1:42069
> request foto.jpg
- Updating nodes who have chunks for file foto.jpg
- File foto.jpg information internally updated.
- File foto.jpg download progress: (10.0%)
- File foto.jpg download progress: (20.1%)
- File foto.jpg download progress: (30.0%)
- ...
- File foto.jpg download progress: (80.0%)
- File foto.jpg download progress: (90.1%)
- File foto.jpg download progress: (100.0%)
- Sent update chunks packet to tracker for file foto.jpg
- File foto.jpg was successfully downloaded in 1.15971392s
```
]

#figure(caption: "Nodo publicador", supplement: "Entidade")[  
```bash
$ ./out/bin/node
- UDP server started on 0.0.0.0:8081
- Connected to tracker on 127.0.0.1:42069
- Type 'help' for a list of available commands
> publish temp/foto.jpg
- Added file foto.jpg to pending files
- Sent publish file packet to tracker
- File foto.jpg published in the network successfully
- Request chunks packet received from 127.0.0.1:8082
- ...
- Request chunks packet received from 127.0.0.1:8082
```
]

== Performance

Testes realizados num computador com processador i5-8300H, 16GB de RAM 2666MHZ DDR4 e SDD.

#table(align: center + horizon, columns: (1fr, 1fr, 1fr), 
[Tamanho de ficheiro], [Número de nodos envolvidos], [Tempo de transferência],
[14MB], [2], [1.15s],
[100MB], [3], [14.27s]
)

= Limitações e Trabalho Futuro <limitations>
Nesta secção iremos nos debruçar sobre algumas limitações que o nosso programa tem e sugerir algum trabalho futuro nesse sentido.

Uma grande limitação deste trabalho é a não existência da funcionalidade de particionar um _chunk_ em pedaços no envio entre _nodes_. Apesar de estarmos a usar #chunk_id_bit_size _bits_ para a identificação de um _chunk_, e com isso, o tamanho do _chunk_ escala bem com o tamanho do ficheiro, o protocolo UDP só consegue enviar 65,535 bytes num só pacote. Como não há qualquer fragmentação de pacotes feita no nível aplicacional, o limite de tamanho de ficheiro passará a ser por volta dos 4.19GB (o tamanho de ficheiro para qual o tamanho da chunk beira o máximo do UDP). Tínhamos planeado desde cedo implementar esta funcionalidade, mas não a conseguimos fazer em tempo útil.

// Uma vez que o tamanho máximo de um pacote UDP é 64 KB e visto que estamos a utilizar #chunk_id_bit_size _bits_ para a codificação dos _chunks_, o tamanho máximo de um ficheiro fica limitado a 4.19 GB. De facto, aumentando o número de _bits_ utilizados para codificação de _chunks_ para, por exemplo, 18 _bits_ permitiria com que fosse possível transferir ficheiros de até 16.78 GB. No entanto, se precisamos de aumentar o número de _bits_ sempre que precisamos de aumentar o tamanho máximo que um ficheiro pode ter, estamos perante um problema. A melhor solução passaria por criar um sistema de _slices_, onde o ficheiro seria dividido em _slices_, e cada _slice_, por sua vez, dividida em _chunks_. Todas as comunicações teriam que ser feitas baseadas em _slices_ e não _chunks_. Assim, o número de _bits_ necessários para codificar as _slices_ passaria a ser constante, pois só com um ficheiro extremamente grande é que existiram mais que $2^16$ _slices_.

O algoritmo de escalonamento foi a parte do trabalho onde mais ficou aquém da nossa expectativa de qualidade. O algoritmo tem pouca robustez a várias situações que podem acontecer em ambientes de sistemas distribuídos e está muito além do ótimo. Estimativas de RTT deviam de ser levadas em conta no cálculo do timeout para cada nodo. Entre os 100 milissegundos de cada chamada ao tick, as chunks pedidas no tick anterior já podem ter sido descarregadas e o nodo fica a "dormir" desnecessariamente até o próximo tick, aumentando o tempo de transferência. 

A escolha da linguagem Golang não ajudou. Também por inexperiência nossa com a linguagem, existem vários _edge cases_ que o código pode ter, devido à natureza de existirem valores nulos, à forma pouco robusta de _handling_ de erros, etc. Também notámos a falta de algumas estruturas de dados que podiam ser úteis e não estão presentes na _standard lib_. Existe a falta de um _standard_ de serialização para binário, que levou a termos que recorrer a _reflection_ para a fazer de forma escalável para as várias estruturas de dados, o que diminuiu a performance do programa em geral. Entre várias outras ergonomias questionáveis da linguagem, chegamos à conclusão que Golang não foi uma boa escolha para o desenvolvimento deste trabalho.

Outro dos pontos em que o programa peca é a existência de apenas um _tracker_. Apesar disto ter sido, de certa forma, imposto pelo enunciado, é fácil perceber que num sistema distribuído como este, existir apenas um e um só nó com este trabalho é uma péssima escolha. Seja porque este pode facilmente falhar, seja porque este não consegue aguentar com tanta carga e dar respostas em tempo útil. Em redes reais de partilha _peer to peer_ em que existem _trackers_, este número seria sempre mais elevado e adequado às características da rede.

= Conclusão

Em jeito de conclusão, considerámos que este foi um dos trabalhos mais interessantes desenvolvidos ao longo da licenciatura, apesar de que os seus maiores desafios envolverem mais temas de sistemas distribuídos, do que propriamente comunicações de computadores. Apesar de não o termos feito com todo o brio que desejavamos, somos da opinião, ainda assim, que temos aqui um projeto bastante positivo e sólido, em que implementamos todas as funcionalidades pedidas.

#pagebreak()

= Anexos

#let config_file_figure = figure.with(supplement: "Anexo de Configuração", numbering: "I", kind: "configfile")

<options>
#config_file_figure(caption: `named.conf.options`)[
  ```
  options {
          directory "/var/cache/bind";
  
          dnssec-validation auto;
  };
  ```
]

<ns1>
#config_file_figure(caption: `ns1.named.conf.local`)[
  ```
  zone "local" {
      type master;
      file "/etc/bind/local.zone";
      allow-transfer { 10.4.4.10; };
      also-notify { 10.4.4.10; };
  };
  
  zone "10.in-addr.arpa" {
      type master;
      file "/etc/bind/reverse.zone";
      allow-transfer { 10.4.4.10; };
      also-notify { 10.4.4.10; };
  };
  ```
]

<ns2>
#config_file_figure(caption: `ns2.named.conf.local`)[
  ```
  zone "local" {
      type slave;
      file "/etc/bind/local.zone";
      masters { 10.4.4.1; };
  };
  
  zone "10.in-addr.arpa" {
      type slave;
      file "/etc/bind/reverse.zone";
      masters { 10.4.4.1; };
  };
  ```
]

<local>
#config_file_figure(caption: `local.zone`)[
  ```
  $ORIGIN local.
  $TTL 1d
  @       IN SOA  ns1.local. admin.local. (
                  2023012301 ; serial
                  8h         ; refresh
                  2h         ; retry
                  4w         ; expire
                  1h         ; minimum
                )
         IN NS   ns1.local.
         IN NS   ns2.local.
  
  ; Define host mappings
  portatil1   IN A    10.1.1.1
  portatil2   IN A    10.1.1.2
  pc1         IN A    10.2.2.1
  pc2         IN A    10.2.2.2
  roma        IN A    10.3.3.1
  paris       IN A    10.3.3.2
  ns1         IN A    10.4.4.1
  ns2         IN A    10.4.4.10
  servidor1   IN A    10.4.4.2
  ```
]

<reverse>
#config_file_figure(caption: `reverse.zone`)[
  ```
  $ORIGIN 10.in-addr.arpa.
  $TTL 1d
  @       IN SOA  ns1.local. admin.local. (
                  2023012301 ; serial
                  8h         ; refresh
                  2h         ; retry
                  4w         ; expire
                  1h         ; minimum
                )
         IN NS   ns1.local.
         IN NS   ns2.local.
  
  ; Define PTR records for reverse DNS
  1.1.1       IN PTR  portatil1.local.
  2.1.1       IN PTR  portatil2.local.
  1.2.2       IN PTR  pc1.local.
  2.2.2       IN PTR  pc2.local.
  1.3.3       IN PTR  roma.local.
  2.3.3       IN PTR  paris.local.
  1.4.4       IN PTR  ns1.local.
  10.4.4      IN PTR  ns2.local.
  2.4.4       IN PTR  servidor1.local.
  ```
]