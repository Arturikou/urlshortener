package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

func Generate(certPath, keyPath string) error {
	{
		cert := &x509.Certificate{
			// указываем уникальный номер сертификата
			SerialNumber: big.NewInt(1658),
			// заполняем базовую информацию о владельце сертификата
			Subject: pkix.Name{
				Organization: []string{"Yandex.Praktikum"},
				Country:      []string{"RU"},
			},
			// разрешаем использование сертификата для 127.0.0.1 и ::1
			IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
			// сертификат верен, начиная со времени создания
			NotBefore: time.Now(),
			// время жизни сертификата — 10 лет
			NotAfter:     time.Now().AddDate(10, 0, 0),
			SubjectKeyId: []byte{1, 2, 3, 4, 6},
			// устанавливаем использование ключа для цифровой подписи,
			// а также клиентской и серверной авторизации
			ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
			KeyUsage:    x509.KeyUsageDigitalSignature,
		}

		// создаём новый приватный RSA-ключ длиной 4096 бит
		// обратите внимание, что для генерации ключа и сертификата
		// используется rand.Reader в качестве источника случайных данных
		privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return fmt.Errorf("failed to generate rsa key: %w", err)
		}

		// создаём сертификат x.509
		certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
		if err != nil {
			return fmt.Errorf("failed to create certificate: %w", err)
		}

		// кодируем сертификат и ключ в формате PEM, который
		// используется для хранения и обмена криптографическими ключами
		var certPEM bytes.Buffer
		err = pem.Encode(&certPEM, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certBytes,
		})
		if err != nil {
			return fmt.Errorf("failed to encode cert: %w", err)
		}

		var privateKeyPEM bytes.Buffer
		err = pem.Encode(&privateKeyPEM, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		})
		if err != nil {
			return fmt.Errorf("failed to encode private key: %w", err)
		}

		if err := os.WriteFile(certPath, certPEM.Bytes(), 0644); err != nil {
			return fmt.Errorf("failed to save cert file: %w", err)
		}

		if err := os.WriteFile(keyPath, privateKeyPEM.Bytes(), 0600); err != nil {
			return fmt.Errorf("failed to save key file: %w", err)
		}

		return nil
	}
}
