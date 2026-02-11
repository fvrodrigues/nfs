#!/bin/bash

printf 'Ao continuar com o uso desse script, você automaticamente concorda com a licença descrita em LICENSE.md
Antes de continuar, deseja ler a licença?\n'
echo "--------------------------------------------------"
printf 'By continuing to use this script, you automatically agree to the license described in LICENSE.md.
Before continuing, would you like to read the license?\n'
echo ""
while true; do
    read -rp "Continue? [y/n]: " resp
    case "$resp" in
        [Yy])
            echo "----------------------------------------------------------"
            cat ../LICENSE.md || { echo "Erro ao mostrar licensa "; exit; }
            echo "-----------/var/lib/nfse-----------------------------------------------"
            break
            ;;
        [Nn])
            echo "Saindo"
            exit
            ;;
        *)
            echo ""
            echo "Digite y ou n."
            echo "Type y or n"
            ;;
    esac
done

echo "Rode esse script como root, após isso, ele irá se autodeletar do computador, pois a permanência dele no sistema é de alto risco"
while true; do
    read -rp "" a
    case "$a" in
        *)
            break
            ;;
    esac
done

groupadd --system nfse 2>/dev/null || true
useradd  --system \
  --gid nfse \
  --home-dir /var/lib/nfse \
  --create-home \
  --shell /usr/bin/nologin \
  --comment "NFSe service user" \
  nfse

echo "Criando env"
install -d -o nfse -g nfse -m 0750 /etc/nfse/
install -o nfse -g nfse -m 0750 /etc/nfse/nfse.env
echo "API_KEY=3f7122b69478a5e4c19fcd92a6b6c583" > /etc/nfse/nfse.env

echo "Criando e configurando pastas necessárias"
install -d -o nfse -g nfse -m 0750 /opt/nfse/
install -d -o nfse -g nfse -m 0750 /opt/nfse/logs/
install -d -o nfse -g nfse -m 0750 /opt/nfse/NFs/
install -o nfse -g nfse -m 0750 /etc/systemd/system/nfse.service

echo "Criando Unit para systemd"
{
echo "[Unit]";
echo "Description=nfse";
echo "After=graphical.target network-online.target";
echo "";
echo "[Service]";
echo "Type=simple";
echo "User=khal";
echo "Group=khal";
echo "";
echo "Environment=\"DISPLAY=:0\"";
echo "Environment=\"XAUTHORITY=/run/user/1000/lyxauth\"";
echo "";
echo "EnvironmentFile=/etc/nfse/nfse.env";
echo "ExecStart=/opt/nfse/nfse";
echo "";
echo "Restart=on-failure";
echo "RestartSec=3";
echo "";
echo "ProtectSystem=strict";
echo "NoNewPrivileges=true";
echo "";
echo "PrivateTmp=true";
echo "ReadWritePaths=/opt/nfse/ -/opt/nfse/NFs/ -/opt/nfse/logs/ -/var/lib/nfse/.cache/rod/browser/";
echo "ExecPaths=-/var/lib/nfse/.cache/rod/browser/ -/var/lib/nfse/.cache/rod/browser";
echo "";
echo "LogsDirectory=nfse";
echo "LogsDirectoryMode=0777";
echo "";
echo "[Install]";
echo "WantedBy=graphical.target";
} > /etc/systemd/system/nfse.service

echo "Unit criada para systemd"

echo "Mudando para o diretório principal"
cd /var/lib/nfse/nfse/cmd/main/ || { echo "Erro ao mudar para o diretório main do projeto"; exit;}

echo "Compilando o código para o diretório que o Systemd gosta"
go build -o /opt/nfse/nfse || { echo "Erro ao compilar código Golang."; exit;}
chown -R nfse:nfse /opt/nfse/nfse

echo "Tudo pronto, reinciando daemon!"

echo "Reiniciando daemon de nfse"
systemctl enable nfse.service
systemctl daemon-reload || { echo "erro em daemon-reload"; exit; }

echo "Reativando serviço"
systemctl reenable nfse.service || { echo "erro em reenable"; exit; }

echo "Reiniciando serviço"
systemctl restart --now nfse.service || {
  echo "erro em restart nfse.service:"; sudo systemctl status nfse.service || { echo "Erro ao tentar ver status do app"; exit; } ;
  }

echo "Agora é hora de assistir o app rodando"
echo "------------------------------------------------------------"
sudo journalctl -u nfse.service -f || { echo "erro em journalctl"; exit; }

