diff_f () {
  diff $1/$2 $1-old/$2
}

cd output/

diff_f blocks block_heights.json
diff_f blocks blocks.json
diff_f blocks txes.json
diff_f blocks txes.log

diff_f contracts addresses.csv
diff_f contracts contract_infos.json
diff_f contracts contract_names.csv
diff_f contracts contract_variables.json
diff_f contracts fungible_balances.json
diff_f contracts fungible_balances_with_leftovers.json
diff_f contracts leftovers-merge.log
diff_f contracts nft_balances.json
diff_f contracts nft_datas.json
diff_f contracts nft_series.json
diff_f contracts stakes.json
diff_f contracts staking_claims.json
diff_f contracts staking_leftovers.json
diff_f contracts staking_master_age.json
diff_f contracts staking_master_claims.json
diff_f contracts token_addresses_and_supplies.json
diff_f contracts token_infos.json
diff_f contracts tokens_list.csv


