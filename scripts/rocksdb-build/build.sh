sudo apt install libsnappy-dev zlib1g-dev libbz2-dev liblz4-dev libzstd-dev

git clone https://github.com/facebook/rocksdb.git
cd rocksdb/
# make all # Makes in debug mode
make shared_lib # Makes required lib in release mode
sudo make install