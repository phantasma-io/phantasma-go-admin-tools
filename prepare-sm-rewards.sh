just build-account-analyzer

./bin/account-analyzer --get-sm-states --address-csv-path=./data/addresses.csv \
  --ignore-address-csv-path=broken-addresses.csv \
  --invalid-address-output-path=./data/invalid-addresses.csv \
  --errors-output-path=./data/errors.log \
  --export-sm-json=./data/sm-rewards.json --verbose
