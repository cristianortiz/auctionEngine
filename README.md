# Online Auction Platform

This project aims to build a web platform for auctioning products like used and damaged vehicles

## Description

The platform will allow authenticated users to participate in real-time vehicle auctions. It will feature sections to view upcoming and past auctions, provide detailed information about each vehicle to be auctioned, and critically, a real-time online auction module that allows bidding in real time.

## Architecture

The project will be developed following a **Modular Monolith** architecture. This means that, while all code will reside in a single codebase and be deployed as a single application (initially), it will be internally organized into well-defined modules, each representing a **Bounded Context** of the business domain.

Within each module, **Domain-Driven Design (DDD)** and **Clean Architecture (Hexagonal Architecture)** principles will be applied to keep the code decoupled, testable, and aligned with business logic.

This approach facilitates understanding and maintenance, and paves the way for a potential future transition to microservices if the project's scale justifies it, as each module is designed to be relatively independent.

## Key Principles

*   **Domain-Driven Design (DDD):** Code is organized around the business domain (Auctions, Vehicle Catalog, Users). Entities, aggregates, and business logic reside in the domain layer.
*   **Clean Architecture / Hexagonal Architecture:** Outer layers (Infrastructure, Interfaces) depend on inner layers (Application, Domain). The domain is the core, agnostic to technology (DB, Web Frameworks, etc.).
*   **Repository Pattern:** Abstracts persistent data access. Domain and application layers interact with repository *interfaces*, whose concrete implementations (e.g., for PostgreSQL) live in the infrastructure layer.
*   **Bounded Contexts:** Each module defines a clear boundary where domain terms and concepts have a specific and consistent meaning.

## Technologies

*   **Backend:** Go (API, business logic, WebSocket server)
*   **Frontend:** React (User interface, consume API and WebSocket)
*   **Database:** PostgreSQL (Relational, for structured and transactional data)
*   **Real-time Communication:** WebSockets (for the online auction between Backend and Frontend)
*   *(Optional/Future):* gRPC (Consider for internal service communication if the monolith is broken down).

## Directory Structure (Modular Monolith in Go)

```
/project-root
    /cmd                # Application entry points (executables)
        /server
            main.go     # Configures, wires up modules, and starts the main server (HTTP/WS)

    /internal           # Private application code (most of the project)
        /shared         # Code shared between modules
            /config         # Global configuration
            /db             # Database connection setup and instance
            /logger         # Logger setup and instance
            /http           # Base HTTP server, common middleware, main router
            /websocket      # The central WebSocket hub, manages generic connections

        # --- Modules (Bounded Contexts) ---

        /auction            # Module/Context: Auction Engine (Online Auction & Auction Logic)
            /domain             # Entities, aggregates, value objects, repository interfaces, domain errors
                auction_lot.go  # Aggregate root (lot state, PlaceBid method)
                bid.go          # Entity/Value Object (an individual bid)
                interfaces.go   # Repository interfaces (AuctionLotRepository, BidRepository)
                errors.go       # Specific errors for the 'auction' domain
                # ... (others if applicable)

            /application        # Use Cases (Interactors), application service interfaces
                place_bid.go        # Use Case: Place a bid
                get_lot_state.go    # Use Case: Get current lot state (for UI and WS initial state)
                handle_timer_tick.go # Use Case: Logic executed by the timer for active lots
                finalize_lot.go     # Use Case: Logic to finalize a lot when time runs out
                join_lot_ws.go      # Use Case: Logic to join a lot's WebSocket (cooperates with shared WS hub)
                service.go          # AuctionService interface (exposes module functionality)

            /infrastructure     # Concrete implementations (DB, WS handlers, HTTP handlers)
                /persistence        # Repository implementations
                    /postgres
                        auction_lot_repository.go # Implements domain.auction.AuctionLotRepository
                        bid_repository.go         # Implements domain.auction.BidRepository

                /websocket          # WebSocket specific logic for this module
                    handlers.go       # WS handlers processing messages and calling use cases
                    messages.go       # JSON message structures for WS communication ('auction' specific)

                /http               # HTTP handlers (e.g., for getting initial state via REST)
                    handlers.go

        /catalog            # Module/Context: Catalog (Vehicle Management & Listing) - **PENDING**
            /domain             # Vehicle entity, VehicleRepository interface
            /application        # Use cases (GetVehicleDetails, ListVehicles), CatalogQueryService interface
            /infrastructure     # Persistence (Postgres VehicleRepository), HTTP handlers
            # ...

        /user               # Module/Context: User / Identity (Authentication & Profiles) - **PENDING**
            /domain             # User entity, UserRepository interface, AuthenticationService interface
            /application        # Use cases (RegisterUser, LoginUser), UserService interface
            /infrastructure     # Persistence (Postgres UserRepository), Authentication implementation, HTTP handlers
            # ...

        # /admin            # Module/Context: Administration - **PENDING**
        # /history          # Module/Context: Auction History - **PENDING**
        # /notification     # Module/Context: Notifications - **PENDING**

    /pkg                # Reusable code that can be imported by other projects (usually empty in monoliths)

    go.mod              # Go dependency management
    go.sum
```

## Initial Development Plan: `auction` Module (Auction Engine)

We will start development by focusing on the most critical module: the **Auction Engine** (`/internal/auction`), which manages the real-time auction logic. This will allow us to validate the real-time technology and address core technical challenges early on.

The initial goal is a *Minimum Viable Product* for auctioning a *single* auction lot.

### Phase 0: Configuration and Shared Foundations

*   Set up the Go project directory structure (`go mod init`).
*   Configure a basic logging system (`shared/logger`).
*   Implement the PostgreSQL database connection (`shared/db`).
*   Set up a basic HTTP server and main router (`shared/http`).
*   Implement the **shared WebSocket Hub** (`shared/websocket/hub.go`) which will manage raw connections and sending messages to client groups (by `lotID`). This hub will contain no business logic.
*   Set up a minimal `user` module (placeholder) with a simple `User` entity (`ID`) and a `UserRepository` interface to allow basic bidder identification. Implement a "fake" or very simple version of the repository in `/infrastructure/persistence/postgres` initially.

### Phase 1: Backend - Core `auction` Module

*   **Domain (`/internal/auction/domain`):**
    *   Model `AuctionLot` as an aggregate root with its attributes (`CurrentPrice`, `EndTime`, `State`, `sync.Mutex`), and the `PlaceBid(userID, amount, minIncrement)` method containing business validation logic and state updates *within the aggregate*. Implement basic timer/time extension logic within domain methods if possible, or at least define the necessary fields (`EndTime`, `LastBidTime`, `TimeExtension`).
    *   Model `Bid` as an entity or value object.
    *   Define the `AuctionLotRepository` and `BidRepository` interfaces.
*   **Infrastructure/Persistence (`/internal/auction/infrastructure/persistence/postgres`):**
    *   Implement `AuctionLotRepository` and `BidRepository` using the shared database connection (`shared/db`). Ensure the use of **database transactions** when saving a valid bid and updating the lot's state to guarantee atomicity.
*   **Application (`/internal/auction/application`):**
    *   Implement the use cases: `PlaceBidUseCase` (orchestrates: loads lot, calls `lot.PlaceBid()`, saves lot/bid via repositories), `GetLotStateUseCase` (loads state/recent bids via repositories), `JoinLotWSUseCase` (orchestrates initial state and registration with the shared WS Hub).
    *   Define and implement the `AuctionService` interface that exposes the necessary methods for the infrastructure (e.g., `PlaceBid(cmd)`, `GetLotState(lotID)`, `ProcessIncomingWSMessage(client, message)`).
*   **Infrastructure/WebSocket (`/internal/auction/infrastructure/websocket`):**
    *   Define the JSON message structures (`messages.go`) for WebSocket communication (e.g., `ClientBidMessage`, `ServerLotUpdateMessage`).
    *   Implement the module-specific `handlers.go` functions called by the `shared/websocket/hub` upon receiving a message for an auction lot. These handlers deserialize the message and call the appropriate application use case (`PlaceBidUseCase`) through the `AuctionService` interface. If the use case call is successful, they notify the `shared/websocket/hub` to broadcast the update (`ServerLotUpdateMessage`).
*   **Wiring (`cmd/server/main.go`):**
    *   Instantiate all components (`shared`, repositories, use cases, `AuctionService`).
    *   Inject dependencies (e.g., use cases receive repositories, `AuctionService` receives use cases, WS handlers receive `AuctionService` and `shared/websocket/hub`).
    *   Configure the HTTP/WS route `/ws/auction/{lotId}` on the main router. The handler for this route gets the `lotId` from the URL, calls `sharedWebsocketHub.RegisterClient` to add the client. When the hub receives a message from this client, it calls a method on the `AuctionService` (e.g., `auctionService.ProcessIncomingWSMessage(client, message)`) for the auction module to handle the message logic.
    *   Mount this WS handler on the main router.
    *   Implement the backend timer: A main goroutine or one within the `shared/websocket/hub` or an `auction` application service (`application/auction/timer_service.go`) that periodically (e.g., every second) iterates over active lots, updates their time state, and asks the `shared/websocket/hub` to broadcast the `ServerLotUpdateMessage`. Implement the "time extension" logic.
    *   Start the HTTP server.

### Phase 2: Frontend - Interface and Real-time Connection

*   Set up the React project.
*   Implement the basic user interface for a single auction lot (display current price, time, recent bids, bidding form). Consider Material Design.
*   Use `fetch` or `axios` to get the initial lot state via a backend REST endpoint (e.g., `/api/v1/auction/lots/{lotId}`).
*   Establish and manage the WebSocket connection to the backend endpoint `/ws/auction/{lotId}`.
*   Implement the logic to send `ClientBidMessage` (JSON) messages via WebSocket when placing a bid.
*   Implement the logic to listen for `ServerLotUpdateMessage` messages from the WebSocket and update the UI state (`useState`, `useReducer`, or state library) to reflect the current price, remaining time, and bid list.
*   Ensure the UI is responsive.

### Phase 3: Integration and Testing

*   Run the backend and frontend.
*   Test the full flow of placing bids from multiple clients simultaneously.
*   Verify real-time synchronization of price and timer.
*   Verify that the time extension logic works correctly.
*   Test the persistence of bids and final state in the database.
*   Debug concurrency and synchronization issues.

### Phase 4: Initial Refinement and Robustness

*   Implement a basic authentication system (even if simplified) to identify bidding users more robustly. This will involve starting minimal development in the `user` module.
*   Improve backend bid validation and frontend error feedback.
*   Implement basic WebSocket reconnection handling.
*   Ensure proper cleanup of resources (goroutines, DB connections, WS connections) when auctions end or the application shuts down.

## Key Technical Challenges

*   **Real-time and Concurrency:** Ensuring auction state consistency and correct synchronization across multiple users and the server, especially under load.
*   **Timer Synchronization:** Implementing the timer logic on the server as the source of truth and keeping it synchronized across all clients, including handling "time extension".
*   **Bid Ordering:** Guaranteeing that bids are processed in the correct order on the backend, even if they arrive nearly simultaneously.
*   **Scalability:** Designing the system to handle an increasing number of concurrent users and active auctions.
*   **Data Consistency:** Ensuring that bids and the final state are persisted atomically in the database.

## Future Modules (Monolith Expansion)

Once the core `auction` module is functional, the other modules can be added to complete the platform:

*   **`catalog`:** Vehicle management (CRUD, photo/document uploads, associating with lots).
*   **`user`:** Full authentication, user profile management (contact details, perhaps payment/billing info).
*   **`admin`:** Administration panel to manage users, vehicles, create and configure auctions/lots, view reports.
*   **`history`:** Viewing past auctions, including bidders and final results per lot.
*   **`notification`:** Notification system (e.g., outbid, auction starting soon).

Each of these modules will be developed following the same internal structure (Domain, Application, Infrastructure) and integrated into the monolith via `main.go`, interacting with other modules through interfaces.

---