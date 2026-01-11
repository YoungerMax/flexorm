# 🌌 FlexORM

**FlexORM** is an **all-in-one database toolkit** that helps you (1) **query your database in any programming language**, (2) **create and run migrations**, and (3) **view your data in a web browser**. It is a **single executable** called `flexorm`.

**⚠️ FlexORM is currently experimental, unstable software undergoing heavily development - it is not (yet) intended for production use cases.**

## ✨ Features

📄 **Define your schema** in plain, simple JSON (with IntelliSense, thanks to JSON schemas).
> **Eliminates** learning custom DSLs and installing extra VSCode extensions.
  
🚀 **Generate portable, lightweight, type-safe clients for any programming language** directly from your schema.
> **Eliminates** illegible type errors, heavy runtimes, slow Rust engines, and language lock-in.

🛡️ **Validate everything** that goes into your database by default.
> **Eliminates** extra data validation libraries and manually writing validation logic.

🔄 **Generate migrations** automatically and apply them to your database.
> **Eliminates** "the library I love doesn't have a migration engine with all the features I need for production."

👀 **View your data** in your web browser with FlexORM Studio.
> **Eliminates** "I have no idea what's in my database."

---

## 🎯 Why FlexORM?

🧰 **All-In-One Toolkit**: Use your database types, manage and run your migrations, validate your data, and generate your database client in a single, unified workflow, letting you ship blazingly fast.

💡 **Single Source of Truth**: One schema drives validation, migrations, and clients across every programming language.

🌍 **True Language Freedom**: Generate native clients for any language and database driver, eliminating being locked into a single ecosystem.

📝 **Simple Schema Format**: Define your database structure in clean, readable JSON instead of learning yet another DSL and installing another VSCode extension.

⚡ **Zero Runtime Overhead**: Emitted clients have no bundled engines or heavy abstractions—just pure, performant native code that uses a database driver of your choosing.

🔒 **Validation by Default**: Automatic validation ensures data integrity before it ever reaches your database, catching errors before they hit your database.

🚫 **Predictable Errors**: Clear, local, and understandable errors—no thousand-line generic type explosions—that ruin your DX. FlexORM generates straightforward, readable clients.

| Feature                   | FlexORM | Prisma | TypeORM | Sequelize | SQLAlchemy | Diesel | Drizzle |
| ------------------------- | ------- | ------ | ------- | --------- | ---------- | ------ | ------- |
| Multi-language support    | ✅       | ❌      | ❌       | ❌         | ❌          | ❌      | ❌       |
| Zero runtime overhead     | ✅       | ❌      | ❌       | ❌         | ❌          | ✅      | ✅       |
| Auto migration generation | ✅       | ✅      | ✅       | ✅         | ✅          | ❌      | ✅       |
| Simple JSON schema        | ✅       | ❌      | ❌       | ❌         | ❌          | ❌      | ❌       |
| Built-in validation       | ✅       | ✅      | ✅       | ✅         | ❌          | ❌      | ⚠️       |
| Type-safe queries         | ✅       | ✅      | ✅       | ❌         | ❌          | ✅      | ✅       |
| No heavy engine/runtime   | ✅       | ❌      | ✅       | ✅         | ✅          | ✅      | ✅       |
| Client code generation    | ✅       | ✅      | ❌       | ❌         | ❌          | ❌      | ❌       |

## 🚀 Roadmap
- [ ] Add support for more programming languages
  - [ ] Python
    - [ ] psycopg2
  - [ ] Go
  - [ ] Ruby
  - [ ] PHP
  - [ ] C#
  - [ ] Java
- [ ] Add support for more database drivers and systems
  - [ ] PostgreSQL
  - [ ] MySQL
  - [ ] SQLite
  - [ ] Serverless databases
- [ ] Expand column types and validation options
  - [ ] boolean
  - [ ] emails (special)
  - [ ] phone numbers (special)
  - [ ] uuid
  - [ ] json
  - [ ] decimal
  - [ ] float
  - [ ] double
  - [ ] bigint
  - [ ] smallint
  - [ ] date
  - [ ] time
  - [ ] interval
  - [ ] bytea
  - [ ] enum
  - [ ] array
  - [ ] pgvector
- [ ] Implement relations support
  - [ ] One-to-many relationships
  - [ ] Many-to-many relationships
- [ ] Enhance FlexORM Studio with advanced features
  - [ ] Data visualization, filtering, and sorting
  - [ ] Edit cells
- [ ] Allow migration scripts to be written in any programming language beyond SQL
- [ ] Support more client functions
  - [ ] delete
  - [ ] update
  - [ ] left join
  - [ ] right join
  - [ ] on conflict