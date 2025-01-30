package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/tebeka/selenium"
)

const port = 4444

func TestLoginPage(t *testing.T) {
	caps := selenium.Capabilities{"browserName": "chrome"}

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", port))
	if err != nil {
		t.Fatalf("Ошибка подключения к ChromeDriver: %v", err)
	}
	defer wd.Quit()

	err = wd.Get("http://localhost:8085/login.html")
	if err != nil {
		t.Fatalf("Ошибка открытия страницы логина: %v", err)
	}

	emailField, err := wd.FindElement(selenium.ByID, "email")
	if err != nil {
		t.Fatalf("Поле email не найдено: %v", err)
	}
	emailField.SendKeys("jk@gmail.com")

	passwordField, err := wd.FindElement(selenium.ByID, "password")
	if err != nil {
		t.Fatalf("Поле password не найдено: %v", err)
	}
	passwordField.SendKeys("123")

	loginButton, err := wd.FindElement(selenium.ByCSSSelector, "button[type='submit']")
	if err != nil {
		t.Fatalf("Кнопка входа не найдена: %v", err)
	}
	loginButton.Click()

	time.Sleep(3 * time.Second)

	userDisplay, err := wd.FindElement(selenium.ByID, "usernameDisplay")
	if err != nil {
		t.Fatalf("Не найден usernameDisplay, логин не выполнен")
	}

	userText, err := userDisplay.Text()
	if err != nil || userText == "" {
		t.Fatalf("Текст usernameDisplay пуст, логин не выполнен")
	}

	t.Log("Тест пройден: Пользователь успешно вошел в систему")
}
