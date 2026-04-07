echo "Tudo pronto, reinciando daemon!"

echo "Reiniciando daemon de nfse"

sudo systemctl enable nfse.service
sudo systemctl daemon-reload || { echo "erro em daemon-reload"; exit; }

echo "Reativando serviço"
sudo systemctl reenable nfse.service || { echo "erro em reenable"; exit; }

echo "Reiniciando serviço"
sudo systemctl restart --now nfse.service || {
  echo "erro em restart nfse.service:"; sudo systemctl status nfse.service || { echo "Erro ao tentar ver status do app"; exit; } ;
  }

echo "Agora é hora de assistir o app rodando"
echo "------------------------------------------------------------"
sudo journalctl -u nfse.service -f || { echo "erro em journalctl"; exit; }
