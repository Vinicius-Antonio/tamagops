# 👾 Tamagops

> **DevOps + Tamagotchi:** Um monitor de infraestrutura gamificado que transforma os recursos do seu sistema operacional em um pet virtual interativo.

---

## 📌 Sobre o Projeto

Ferramentas tradicionais de monitoramento de sistema (como *htop*, *Task Manager* ou *Grafana*) são eficientes, porém frias e estáticas. O **Tamagops** revoluciona a experiência de manutenção e telemetria de infraestrutura gamificando tarefas rotineiras de DevOps.

A saúde, o humor e a evolução do seu companheiro virtual estão estritamente ligados em tempo real ao comportamento do hardware e do sistema operacional da máquina host:

* 🟢 **Uso de CPU Equilibrado (20% ~ 60%):** O pet se mantém alimentado, feliz e acumulando pontos de experiência (**XP**) para evoluir de estágio.
* 🟡 **Sobrecarga de RAM ou Disco Cheio (> 85%):** O ambiente do pet acumula "sujeira digital". O usuário precisa acionar uma rotina de limpeza de cache ou fechar abas/aplicativos ociosos para dar um banho na criatura.
* 🔴 **Processos Travados ou Zumbis (CPU > 90% contínua):** O pet adoece e perde pontos de vida (**HP**). Para curá-lo, o desenvolvedor precisa identificar o processo agressor e aplicar o "remédio" (finalizar o PID culpado via CLI ou interface).

---

## 🏗️ Arquitetura e Stack Tecnológica

O projeto adota uma **Arquitetura Desacoplada (Monorepo)**, separando rigorosamente o processamento de baixo nível da camada de apresentação:

### 🛠️ Backend / Daemon (Território A)
O motor de monitoramento roda em segundo plano sem interface gráfica (headless).
* **Linguagem:** [Go (Golang)](https://golang.org/) — Escolhido por seu baixo consumo de memória RAM (< 15MB), alta performance e concorrência nativa (*Goroutines*).
* **Coleta de Telemetria:** Biblioteca `gopsutil` para leitura multiplataforma de hardware (Linux, Windows e macOS).
* **Persistência:** **SQLite** local (banco de dados embarcado) para registrar o histórico biológico do pet e o "Diário de Bordo".
* **Comunicação:** Servidor HTTP REST (comandos de ação) e **WebSockets** (streaming de dados em milissegundos).

### 🎨 Frontend / UI & Game (Território B)
A interface visual responsiva onde o pet e os gráficos ganham vida.
* **Core:** **React** com **TypeScript** e **Vite** para máxima velocidade de build e segurança de tipagem no consumo do contrato JSON.
* **Desktop Runtime:** **Tauri** ou **Electron** para empacotar a aplicação web como um software desktop nativo leve que roda na bandeja do sistema (*System Tray*).
* **Estilização & Animações:** Animações reativas baseadas em estados biológicos (`HAPPY`, `SICK`, `DIRTY`, `SLEEPING`).

---

## 📂 Estrutura do Repositório

```text
tamagops/
├── daemon/                       # 🛠️ Backend em Go (Serviço de Monitoramento)
│   ├── cmd/
│   │   └── main.go               # Ponto de entrada e inicialização do servidor
│   ├── pkg/
│   │   ├── collector/            # Sensor do SO (CPU, RAM, Disco e Processos)
│   │   ├── engine/               # Máquina de Estados Biológica (Regras do Pet)
│   │   ├── server/               # Handlers HTTP REST e servidor WebSocket
│   │   └── storage/              # Repositório de dados SQLite local
│   ├── go.mod
│   └── go.sum
├── app/                          # 🎨 Frontend em React + TypeScript (UI/UX)
│   ├── src/
│   │   ├── components/           # Componentes visuais (Painéis DevOps, Barras HP/XP)
│   │   ├── services/             # Cliente WebSocket e integradores REST
│   │   ├── sprites/              # Artes, ilustrações e animações do Pet
│   │   ├── types/                # Contratos e interfaces TypeScript
│   │   └── App.tsx
│   ├── package.json
│   └── tsconfig.json
├── docs/
│   └── ARCHITECTURE.md           # 📄 Documentação profunda da arquitetura e JSON
├── .gitignore
└── README.md
```

---

## 🚀 Como Executar Localmente

### Pré-requisitos

* [Go](https://go.dev/dl/) instalado (versão 1.20 ou superior).
* [Node.js](https://nodejs.org/) instalado (versão 18 ou superior) e gerenciador de pacotes (`npm` ou `pnpm`).

### 1. Inicializando o Daemon (Backend Go)

O daemon iniciará os sensores do sistema e abrirá o servidor de comunicação na porta `8585`.

```bash
# Entre na pasta do backend
cd daemon

# Baixe as dependências do módulo
go mod tidy

# Execute o servidor principal
go run cmd/main.go
```

> *Servidor ativo! Teste a comunicação acessando `http://localhost:8585/status` em seu navegador.*

### 2. Inicializando a Interface Visual (Frontend React)

Em um **novo terminal**, inicie a aplicação visual para conectar aos sensores do daemon:

```bash
# Entre na pasta do frontend
cd app

# Instale as dependências do projeto
npm install

# Inicie o servidor de desenvolvimento
npm run dev
```

> *Acesse o link gerado no terminal (geralmente `http://localhost:5173`) para interagir com o Tamagops!*

---

## 🗺️ Roadmap de Evolução

* [x] **Fase 1:** Estrutura do Monorepo e comunicação HTTP básica (REST / Polling).
* [ ] **Fase 2:** Implementação do motor de WebSockets para streaming contínuo em tempo real.
* [ ] **Fase 3:** Integração com SQLite para salvar o progresso de evolução e nível do pet.
* [ ] **Fase 4:** Adição de ações interativas (botões de "Dar Banho" e "Matar Processo" na UI).
* [ ] **Fase 5:** Empacotamento final em binário desktop multiplataforma com **Tauri**.

---

## 👥 Autores & Divisão de Trabalho

Este projeto foi construído em colaboração com separação de responsabilidades no monorepo:

* **Pessoa A (Backend & Concorrência):** Desenho da arquitetura em Go, sensores do sistema operacional, concorrência com Goroutines e lógica da máquina de estados biológica. — [Vinicius-Antonio](https://github.com/Vinicius-Antonio)
* **Pessoa B (Frontend & UI/UX):** Interface visual em React/TypeScript, cliente WebSocket, renderização de arte, gamificação de painéis DevOps e empacotamento desktop. — [reimuh](https://github.com/reimuh)
