import requests
from web3.auto import w3
from eth_account.messages import encode_defunct, _hash_eip191_message

class Transaction:
    def __init__(self, dapp, chain_id, action, from_, fee, fee_recipient, nonce, version, params):
        self.dapp = dapp
        self.chain_id = chain_id
        self.action = action
        self.from_ = from_
        self.fee = fee
        self.fee_recipient = fee_recipient
        self.nonce = nonce
        self.version = version
        self.params = params
        self.sig = ''
        self.hash = self.get_hash()

    def __str__(self):
        return 'dapp:' + self.dapp + '\n' + \
            'chainID:' + self.chain_id + '\n' + \
            'action:' + self.action + '\n' + \
            'from:' + self.from_ + '\n' + \
            'fee:' + self.fee + '\n' + \
            'feeRecipient:' + self.fee_recipient + '\n' + \
            'nonce:' + self.nonce + '\n' + \
            'version:' + self.version + '\n' + \
            'params:' + self.params + '\n'

    def get_hash(self):
        message = encode_defunct(text=str(self))
        message_hash = _hash_eip191_message(message)
        return w3.to_hex(message_hash)

    def to_dict(self):
        return {'dapp': self.dapp,
            'chainID': self.chain_id,
            'action': self.action,
            'from':  self.from_,
            'fee': self.fee,
            'feeRecipient': self.fee_recipient,
            'nonce': self.nonce,
            'version': self.version,
            'params': self.params,
            'sig': self.sig}

    def sign(self, signer):
        sig = signer.sign(str(self))
        self.sig = sig
        return sig
    
    def post(self, submit_url):
        return requests.post(submit_url, json=self.to_dict())