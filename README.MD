# API Server for allo.li

Runned on server as service

Commands:
- service api-allo.li start
- service api-allo.li restart
- service api-allo.li stop

For reload daemon after change .service file run:
- systemctl daemon-reload


Some systemctl commands:
- systemctl list-unit-files | grep enabled


## FIX: Одинаковые названия полей во вложенных структурах
В sqlx одинаковые ключи полей вложенных структур опускаются:

>Row headers can be ambiguous (SELECT 1 AS a, 2 AS a), and the result of Columns() does not fully qualify column names in queries like:
>`SELECT a.id, a.name, b.id, b.name FROM foos AS a JOIN foos AS b ON a.parent = b.id;`
>making a struct or map destination ambiguous. Use AS in your queries to give columns distinct names, rows.Scan to scan them manually, or SliceScan to get a slice of results.

Решение: В адаптере включить вывод полей с префиксом
https://github.com/go-sql-driver/mysql#columnswithalias
Затем, переименновать поля в структурах из `db:"type"` в `db:"b.type"`