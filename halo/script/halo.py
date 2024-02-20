import json, argparse, time, sys
import requests
from decimal import Decimal
import everpay
from colorama import Fore, Style
from transaction import Transaction

def get_singer(fn):
    try:
        json.load(open(fn))
    except json.JSONDecodeError:
        return everpay.ETHSigner(open(fn).read().strip())
    else:
        return everpay.ARSigner(fn)

halo_decimals = 18

parser = argparse.ArgumentParser(description='A halo cmd client')
parser.add_argument('-w', '--wallet', type=str, dest='wallet', help='wallet file contain eth/ar private key')
parser.add_argument('-r', '--router', type=str, dest='router', help='halo router url to submit tx')

subparsers = parser.add_subparsers(dest='action', help='halo action help')

parser_transfer = subparsers.add_parser('transfer', help='transfer halo')
parser_transfer.add_argument('-t', '--to', type=str, dest='to', help='Who to transfer to')
parser_transfer.add_argument('-a', '--amount', type=str, dest='amount', help='amount to transfer')
parser_transfer.add_argument('-r', '--raw_amount', action='store_true', dest='raw', help='raw amount to transfer, not multiply 10^decimals')

parser_unstake = subparsers.add_parser('unstake', help='unstake halo')
parser_unstake.add_argument('-p', '--pool', type=str, dest='pool', help='which pool to unstake')
parser_unstake.add_argument('-a', '--amount', type=str, dest='amount', help='amount to unstake')

parser_stake = subparsers.add_parser('stake', help='stake halo')
parser_stake.add_argument('-p', '--pool', type=str, dest='pool', help='which pool to stake')
parser_stake.add_argument('-a', '--amount', type=str, dest='amount', help='amount to stake')

parser_propose = subparsers.add_parser('propose', help='propose')
parser_propose.add_argument('-n', '--name', type=str, dest='name', help='proposal name')
parser_propose.add_argument('-c', '--code', type=str, dest='code', help='source code file of proposal')
parser_propose.add_argument('-d', '--data', type=str, dest='data', help='initial data file of proposal')
parser_propose.add_argument('-t', '--times', type=int, dest='times', help='run times of proposal')
parser_propose.add_argument('-s', '--start', type=int, dest='start', help='start times of proposal')
parser_propose.add_argument('-e', '--end', type=int, dest='end', help='end times of proposal')
# todo: TxActionsSupported

args = parser.parse_args()

signer = get_singer(args.wallet)
info = requests.get(args.router + '/info').json()
dapp = info['dapp']
chain_id = info['chainID']
fee_recipient = info['feeRecipient']
submit_url = args.router + '/submit'

if args.action == 'transfer':
    if not args.to or not args.amount:
        print(Fore.RED + 'invalid transfer options' + Style.RESET_ALL)
        sys.exit(1)
    if args.raw:
        amount = args.amount
    else:
        amount = str(int(Decimal(args.amount) * 10**halo_decimals))
    params = {
        'to': args.to,
        'amount': amount,
    }

elif args.action == 'unstake' or args.action == 'stake':
    if not args.pool or not args.amount:
        print(Fore.RED + 'invalid unstake/stake options' + Style.RESET_ALL)
        sys.exit(1)

    amount = str(int(Decimal(args.amount) * 10**halo_decimals))
    params = {
        'stakePool': args.pool,
        'amount': amount,
    }

elif args.action == 'propose':
    if not args.name or not args.code:
        print(Fore.RED + 'invalid propose options' + Style.RESET_ALL)
        sys.exit(1)
    
    if not args.times and not args.start:
        print(Fore.RED + 'invalid propose options' + Style.RESET_ALL)
        sys.exit(1)

    code = open(args.code).read()
    initData = ''
    if args.data:
        initData = open(args.data).read()

    params = {
        'name':	args.name,
        'source': code,
        'start': args.start,
        'end': args.end,
        'initData': initData,
        #'onlyAcceptedTxActions': ['call'],
    }
    if args.times:
        params['runTimes'] = args.times
    else:
        params['start'] = args.start
        params['end'] = args.end
else:
    parser.print_help()

tx = Transaction(
    dapp = dapp,
    chain_id = chain_id,
    action = args.action,
    from_ = signer.address,
    fee = '0',
    fee_recipient= fee_recipient,
    nonce= str(int(time.time() * 1000)),
    version= 'v1',
    params= json.dumps(params)
)
tx.sign(signer)
result = tx.post(submit_url)
print('sumbit tx return:', result.content)