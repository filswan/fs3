//web3.js
const contractAddress = ''
const contractABI=[]
var contract=null;
function Init(callback){
  if (typeof window.ethereum === "undefined") {
    // alert("Looks like you need a Dapp browser to get started.");
    alert("Consider installing MetaMask!");
  } else {
    ethereum.enable()
      .catch(function (reason) {
        if (reason === "User rejected provider access") {
        } else {
          alert("There was an issue signing you in.");
        }
      }).then(function (accounts) {
      var currentProvider = web3.currentProvider;
      var Web3 = web3js.getWeb3();
      web3 = new Web3();
      web3.setProvider(currentProvider);
      contract = new web3.eth.Contract(contractABI, contractAddress);
      callback(accounts[0]);
      sessionStorage.removeItem('addrWeb')
      sessionStorage.setItem('addrWeb', accounts[0])
    });
  }
}

export default {
  Init
}
