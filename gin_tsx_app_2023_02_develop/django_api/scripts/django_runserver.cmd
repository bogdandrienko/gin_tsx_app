@echo OFF

cd ..

call .\env\Scripts\activate.bat



python manage.py runserver 0.0.0.0:8084



cmd