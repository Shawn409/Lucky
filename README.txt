CareU Setup Guide
=================

This guide explains how to run the CareU project on another computer.


1. Requirements
---------------

Install these first:

1. XAMPP, or any local MySQL/MariaDB server
2. Go programming language
3. A browser such as Chrome, Edge, or Firefox


2. Project Files
----------------

Important files in this folder:

- main.go                 Backend server and API
- sihatai-sarawak.html    CareU web interface
- careu.sql               MySQL database export for import
- go.mod / go.sum         Go dependencies
- .env.example            Example database settings

Do not open sihatai-sarawak.html directly if you want login/register/database
features to work. The website must be opened through the Go server.


3. Import Database Using phpMyAdmin
-----------------------------------

1. Open XAMPP Control Panel.
2. Start Apache.
3. Start MySQL.
4. Open phpMyAdmin:

   http://localhost/phpmyadmin

5. Click "New" on the left side.
6. Create a database named:

   careu

7. Click the careu database.
8. Click "Import".
9. Choose the file:

   careu.sql

10. Click "Go" to import.

If the import fails because tables already exist:

1. Drop/delete the old careu database.
2. Create careu again.
3. Import careu.sql again.


4. Run The Website
------------------

Open PowerShell in the project folder.

Example:

   cd C:\Projects\Lucky

Then run:

   go run .

If it runs correctly, you should see something like:

   CareU running at http://localhost:8080

Open this URL in your browser:

   http://localhost:8080


5. Default Database Settings
----------------------------

The project uses these default settings:

   Database host: 127.0.0.1:3306
   Database name: careu
   Database user: root
   Database pass: empty password

This matches the common XAMPP MySQL default.

If your MySQL root account has a password, set it before running the server:

   $env:CAREU_DB_PASS="your_mysql_password"
   go run .

If your MySQL uses another port:

   $env:CAREU_DB_ADDR="127.0.0.1:3307"
   go run .


6. How To Use CareU
-------------------

1. Open http://localhost:8080
2. Login using the demo account, or register a new account.
3. Complete or update the Profile page.
4. Add health readings in Daily Entry.
5. Check Dashboard, Health Summary, Mental Health, and Weekly Trend.

Demo login account after importing careu.sql:

   Email: admin@gmail.com
   Password: admin123

Imported demo data may include old accounts. If you do not know the password,
just register a new account.


7. Common Problems
------------------

Problem: Login/Register does not work.
Solution:
- Make sure you opened http://localhost:8080, not the HTML file directly.
- Make sure MySQL is running in XAMPP.
- Make sure the careu database exists.

Problem: "database unavailable" or connection error.
Solution:
- Check whether MySQL is started.
- Check database name is careu.
- Check username/password in the environment variables.

Problem: Port 8080 is already used.
Solution:
Run the app on another port:

   $env:CAREU_ADDR=":8081"
   go run .

Then open:

   http://localhost:8081

Problem: Go command is not recognized.
Solution:
- Install Go from https://go.dev/dl/
- Restart PowerShell after installing Go.


8. Notes
--------

CareU is a prototype for personal health monitoring and review. It is not a
replacement for professional medical diagnosis or treatment.
