# PostgreSQL va pgAdmin

Docker Desktop ishlayotgan bo'lishi kerak. Compose `.env` dagi `DB_NAME`,
`DB_USER`, `DB_PASSWORD` va `DB_PORT` qiymatlarini ishlatadi.

```powershell
docker compose up -d
docker compose ps
```

Go dasturi kompyuterda ishlasa, `DB_HOST=localhost` bo'lsin.
Migratsiyalar Go dasturidagi `config.DBConnect()` orqali bajariladi:

```powershell
make run
```

pgAdmin: http://localhost:5050

- Email: `.env` dagi `PGADMIN_EMAIL` (standart: `admin@example.com`).
- Parol: `.env` dagi `PGADMIN_PASSWORD`.

pgAdmin ichida **Register > Server** ni tanlang. **General > Name**:
`Task Manager`. **Connection** bo'limida:

| Maydon | Qiymat |
| --- | --- |
| Host name/address | `postgres` |
| Port | `5432` |
| Maintenance database | `.env` dagi `DB_NAME` |
| Username | `.env` dagi `DB_USER` |
| Password | `.env` dagi `DB_PASSWORD` |

Konteynerlarni to'xtatish:

```powershell
docker compose down
```

Ma'lumotlar Docker volume'larda saqlanadi. `docker compose down -v`
volume'larni va saqlangan ma'lumotlarni o'chiradi.

PostgreSQL birinchi ishga tushganda DB va foydalanuvchini yaratadi.
Keyinchalik `.env` dagi DB parolini almashtirish mavjud DB parolini
avtomatik o'zgartirmaydi.
