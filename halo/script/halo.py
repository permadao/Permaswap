#!/usr/bin/env python

import json, time, sys
import requests
from decimal import Decimal
from optparse import OptionParser 
from colorama import Fore, Style
import everpay
from transaction import Transaction

def get_singer(fn):
    try:
        json.load(open(fn))
    except json.JSONDecodeError:
        return everpay.ETHSigner(open(fn).read().strip())
    else:
        return everpay.ARSigner(fn)
    
halo_decimals = 18

parser = OptionParser()  
parser.add_option('-r', '--router',  dest='router', default='http://127.0.0.1:8080',
                  help='router to submit tx')
parser.add_option('-w', '--wallet',  dest='wallet', 
                  help='eth/ar wallet file')

parser.add_option('-c', '--action', dest='action',  
                  help='tx action: transfer, stake, unstake, join, leave, propose, call')
parser.add_option('-t', '--to',  dest='to',
                  help='receiver')
parser.add_option('-a', '--amount',  dest='amount', 
                  help='amount to transfer or stake')
parser.add_option('-p', '--pool',  dest='pool', 
                  help='pool to stake or unstake')

(options, args) = parser.parse_args()  

signer = get_singer(options.wallet)

info = requests.get(options.router + '/info').json()
dapp = info['dapp']
chain_id = info['chainID']
fee_recipient = info['feeRecipient']
submit_url = options.router + '/submit'

if options.action == 'transfer':
    if not options.to or not options.amount:
        print(Fore.RED + 'invalid transfer options' + Style.RESET_ALL)
        sys.exit(1)
    
    amount = str(Decimal(options.amount) * 10**halo_decimals)
    params = {
        'to': options.to,
        'amount': amount,
    }

elif options.action == 'unstake' or options.action == 'stake':
    if not options.pool or not options.amount:
        print(Fore.RED + 'invalid unstake/stake options' + Style.RESET_ALL)
        sys.exit(1)

    amount = str(Decimal(options.amount) * 10**halo_decimals)
    params = {
        'stakePool': options.pool,
        'amount': amount,
    }
    
else:
    print(Fore.RED + 'invalid action' + Style.RESET_ALL)
    sys.exit(1)


tx = Transaction(
    dapp = dapp,
    chain_id = chain_id,
    action = options.action,
    from_ = signer.address,
    fee = '0',
    fee_recipient= fee_recipient,
    nonce= str(int(time.time() * 1000)),
    version= 'v1',
    params= json.dumps(params)
)
tx.sign(signer)
result = tx.post(options.router + '/submit')
print('sumbit tx return:', result.content)