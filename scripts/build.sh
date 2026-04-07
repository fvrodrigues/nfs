#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR" || exit
echo "" > vazio.txt

printf 'Ao continuar com o uso desse script, você automaticamente concorda com a licença descrita em LICENSE.md'
echo "--------------------------------------------------"
printf 'By continuing to use this script, you automatically agree to the license described in LICENSE.md.'

echo "----------------------------------------------------------"
printf 'Copyright (c) 2026 Khaled

        Todos os direitos reservados.

        Este software é de natureza proprietária e confidencial.

        A permissão concedida limita-se exclusivamente ao uso do software,
        mediante contrato comercial válido ou autorização expressa e por escrito
        do titular dos direitos autorais.

        É vedada, sem autorização prévia e por escrito do titular,
        a reprodução, cópia, modificação, adaptação, distribuição,
        sublicenciamento, engenharia reversa, descompilação,
        revenda ou qualquer outra forma de exploração do software.

        A aquisição do software ou a contratação de serviços relacionados
        não implica, necessariamente, na transferência de titularidade,
        acesso ao código-fonte, direito de redistribuição,
        modificação ou qualquer outro direito além do uso autorizado.

        ---

        Copyright (c) 2026 Khaled

        All rights reserved.

        This software is proprietary and confidential.

        Permission is granted only to use the software
        under a valid commercial agreement or written authorization
        from the copyright holder.

        No part of this software may be copied, modified, distributed,
        sublicensed, reverse engineered, decompiled, resold,
        or otherwise exploited without explicit written permission.

        Purchasing or contracting services involving this software
        does not necessarily grant ownership rights, source code access,
        redistribution rights, or modification rights.\n'
echo "----------------------------------------------------------"

while true; do
    echo "Continue and agree with the abovementioned terms and license? y: yes n: no  [y/n]: "
    read -rp "Continuar e concordar com os termos da licença acima? y: sim n: não  [y/n]:" resp
    case "$resp" in
        [Yy])
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
chown -R nfse:nfse ./nfse
chown nfse:nfse ~/daemon.sh
chmod 750 ~/daemon.sh


go version || {
  echo "instale o Golang primeiro";
  exit;
}

groupadd nfse 2>/dev/null || true && \
useradd -m -g nfse -s /usr/bin/bash -c "NFSe user" nfse && \
echo "nfse:nfse123" | sudo chpasswd

echo "Criando env"
install -d -o nfse -g nfse -m 0750 /etc/nfse/
install -o nfse -g nfse -m 0750./vazio.txt /etc/nfse/nfse.env
echo "API_KEY=3f7122b69478a5e4c19fcd92a6b6c583" > /etc/nfse/nfse.env

echo "Criando e configurando pastas necessárias"
install -d -o nfse -g nfse -m 0750 /opt/nfse/
install -d -o nfse -g nfse -m 0750 /opt/nfse/logs/
install -d -o nfse -g nfse -m 0750 /opt/nfse/NFs/
install -o nfse -g nfse -m 0750 ./vazio.txt /etc/systemd/system/nfse.service

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
echo "Environment=\"DISPLAY=$DISPLAY\"";
echo "Environment=\"XAUTHORITY=$XAUTHORITY\"";
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
echo "ExecPaths=-/opt/nfse -/opt/nfse/nfse -/var/lib/nfse/.cache/rod/browser/ -/var/lib/nfse/.cache/rod/browser";
echo "";
echo "LogsDirectory=nfse";
echo "LogsDirectoryMode=0777";
echo "";
echo "[Install]";
echo "WantedBy=graphical.target";
} > /etc/systemd/system/nfse.service

echo "Unit criada para systemd"

echo "Mudando para o diretório principal"
rm -rf /var/lib/nfse/nfse
cp -R ./nfse/ /var/lib/nfse/
cd /var/lib/nfse/nfse/cmd/main/ || { echo "Erro ao mudar para o diretório main do projeto"; exit;}

echo "Compilando o código para o diretório que o Systemd gosta"
go build -o /opt/nfse/nfse || { echo "Erro ao compilar código Golang."; exit;}
chown -R nfse:nfse /opt/nfse/nfse
chown nfse:nfse ~/daemon.sh