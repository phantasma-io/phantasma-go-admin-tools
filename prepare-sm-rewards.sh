just build-account-analyzer

./bin/account-analyzer --get-sm-states --address-csv-path=./data/addresses.csv \
  --ignore-address-csv-path=./data/ignore-addresses.csv \
  --invalid-address-output-path=./data/invalid-addresses.csv \
  --errors-output-path=./data/errors.log \
  --sm-states-file-path=./data/sm-states.json \
  --export-sm-json=./data/sm-rewards.json --verbose
