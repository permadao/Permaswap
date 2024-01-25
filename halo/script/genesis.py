import everpay, json, os

halo_address = '0xD110107aDb30BCe6C0646EAF77cC1C815012331d'
pk = os.getenv('pk')
signer = everpay.ETHSigner(pk)
api_server = 'https://api.everpay.io'

genesis_params = {
    'dapp': 'halo',
    'chainID': "1",
    'govern': '0x95Eb44B81d992534c86994df7D25f5bebE285057',
    'feeRecipient': '0xc6B2FcadaEC9FdC6dA8e416B682d4915F85986f6',
    'routerMinStake': "80000000000000000000000",
    'routers': [
        '0xD110107aDb30BCe6C0646EAF77cC1C815012331d'
    ],
    'routerStates':{
        '0xD110107aDb30BCe6C0646EAF77cC1C815012331d':{
            'router': '0xD110107aDb30BCe6C0646EAF77cC1C815012331d',
            'swapFeeRecipient': '0xc6B2FcadaEC9FdC6dA8e416B682d4915F85986f6',
            'pools':{
                '0x0750e26dbffb85e66891b71b9e1049c4be6d94dab938bbb06573ca6178615981':{
                    'tokenXTag': 'arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543',
                    'tokenYTag': 'ethereum-usdc-0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48',
                    'feeRatio': '0.003'
                },
                '0x5ac5d3598820e140cf5829cd6e50ade648d94496da540c51c9a19f11e06daae8': {
                    'tokenXTag': 'ethereum-eth-0x0000000000000000000000000000000000000000',
                    'tokenYTag': 'ethereum-map-0x9e976f211daea0d652912ab99b0dc21a7fd728e4',
                    'feeRatio': '0.003'
                },
                '0x5f0ac5160cd3c105f9db2fb3e981907588d810040eb30d77364affd6f4435933': {
                    'tokenXTag': 'arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543',
                    'tokenYTag': 'ethereum-eth-0x0000000000000000000000000000000000000000',
                    'feeRatio': '0.003'
                },
                '0x6e80137a5bbb6ae6b683fcd8a20978d6b4632dddc78aa61945adbcc5a197ca0f': {
                    'tokenXTag': 'arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543',
                    'tokenYTag': 'ethereum-ans-0x937efa4a5ff9d65785691b70a1136aaf8ada7e62',
                    'feeRatio': '0.003'
                },
                '0x7200199c193c97012893fd103c56307e44434322439ece7711f28a8c3512c082': {
                    'tokenXTag': 'ethereum-usdc-0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48',
                    'tokenYTag': 'everpay-acnh-0x72247989079da354c9f0a6886b965bcc86550f8a',
                    'feeRatio': '0.0005'
                },
                '0x7eb07d77601f28549ff623ad42a24b4ac3f0e73af0df3d76446fb299ec375dd5': {
                    'tokenXTag': 'ethereum-eth-0x0000000000000000000000000000000000000000',
                    'tokenYTag': 'ethereum-usdc-0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48',
                    'feeRatio': '0.003'
                },
                '0x94170544e7e25b6fc216eb044c1c283c89781bfb92bfeda3054488497bd654b6': {
                    'tokenXTag': 'arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543',
                    'tokenYTag': 'arweave-stamp-TlqASNDLA1Uh8yFiH-BzR_1FDag4s735F3PoUFEv2Mo',
                    'feeRatio': '0.003'
                },
                '0xbb546a762e7d5f24549cfd97dfa394404790293277658e42732ab3b2c4345fa3': {
                    'tokenXTag': 'arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543',
                    'tokenYTag': 'arweave-ardrive--8A6RexFkpfWwuyVO98wzSFZh0d6VJuI-buTJvlwOJQ',
                    'feeRatio': '0.003'
                },
                '0xdb7b3480f2d1f7bbe91ee3610664756b91bbe0744bc319db2f0b89efdf552064': {
                    'tokenXTag': 'ethereum-usdc-0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48',
                    'tokenYTag': 'ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7',
                    'feeRatio': '0.0005'
                },
                '0xdc13faadbd1efdaeb764f5515b20d88c5b9fa0c507c0717c7013b1725e398717': {
                    'tokenXTag': 'arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543',
                    'tokenYTag': 'arweave-u-KTzTXT_ANmF84fWEKHzWURD1LWd9QaFR9yfYUwH2Lxw',
                    'feeRatio': '0.003'
                },
                '0xfed43e1896acd068aed61086b4b038bed6785eb8c29fdba242ffe259e6f9581f': {
                    'tokenXTag': 'bsc-usdt-0x55d398326f99059ff775485246999027b3197955',
                    'tokenYTag': 'ethereum-usdc-0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48',
                    'feeRatio': '0.0005'
                }
            },
            'everTokens':{
                'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543': {
                    'id': 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543',
                    'tag': 'arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543',
                    'symbol': 'AR',
                    'decimals': 12,
                    'chainType': 'arweave,ethereum',
                    'chainID': '0,1'
                },
                '-8A6RexFkpfWwuyVO98wzSFZh0d6VJuI-buTJvlwOJQ': {
                    'id': '-8A6RexFkpfWwuyVO98wzSFZh0d6VJuI-buTJvlwOJQ',
                    'tag': 'arweave-ardrive--8A6RexFkpfWwuyVO98wzSFZh0d6VJuI-buTJvlwOJQ',
                    'symbol': 'ARDRIVE',
                    'decimals': 18,
                    'chainType': 'arweave',
                    'chainID': '0'
                },
                'TlqASNDLA1Uh8yFiH-BzR_1FDag4s735F3PoUFEv2Mo': {
                    'id': 'TlqASNDLA1Uh8yFiH-BzR_1FDag4s735F3PoUFEv2Mo',
                    'tag': 'arweave-stamp-TlqASNDLA1Uh8yFiH-BzR_1FDag4s735F3PoUFEv2Mo',
                    'symbol': 'STAMP',
                    'decimals': 12,
                    'chainType': 'arweave',
                    'chainID': '0'
                },
                'KTzTXT_ANmF84fWEKHzWURD1LWd9QaFR9yfYUwH2Lxw': {
                    'id': 'KTzTXT_ANmF84fWEKHzWURD1LWd9QaFR9yfYUwH2Lxw',
                    'tag': 'arweave-u-KTzTXT_ANmF84fWEKHzWURD1LWd9QaFR9yfYUwH2Lxw',
                    'symbol': 'U',
                    'decimals': 6,
                    'chainType': 'arweave',
                    'chainID': '0'
                },
                '0x55d398326f99059ff775485246999027b3197955': {
                    'id': '0x55d398326f99059ff775485246999027b3197955',
                    'tag': 'bsc-usdt-0x55d398326f99059ff775485246999027b3197955',
                    'symbol': 'USDT',
                    'decimals': 18,
                    'chainType': 'bsc',
                    'chainID': '56'
                },
                '0x937efa4a5ff9d65785691b70a1136aaf8ada7e62': {
                    'id': '0x937efa4a5ff9d65785691b70a1136aaf8ada7e62',
                    'tag': 'ethereum-ans-0x937efa4a5ff9d65785691b70a1136aaf8ada7e62',
                    'symbol': 'ANS',
                    'decimals': 18,
                    'chainType': 'ethereum',
                    'chainID': '1'
                },
                '0x0000000000000000000000000000000000000000': {
                    'id': '0x0000000000000000000000000000000000000000',
                    'tag': 'ethereum-eth-0x0000000000000000000000000000000000000000',
                    'symbol': 'ETH',
                    'decimals': 18,
                    'chainType': 'ethereum',
                    'chainID': '1'
                },
                '0x9e976f211daea0d652912ab99b0dc21a7fd728e4': {
                    'id': '0x9e976f211daea0d652912ab99b0dc21a7fd728e4',
                    'tag': 'ethereum-map-0x9e976f211daea0d652912ab99b0dc21a7fd728e4',
                    'symbol': 'MAP',
                    'decimals': 18,
                    'chainType': 'ethereum',
                    'chainID': '1'
                },
                '0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48': {
                    'id': '0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48',
                    'tag': 'ethereum-usdc-0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48',
                    'symbol': 'USDC',
                    'decimals': 6,
                    'chainType': 'ethereum',
                    'chainID': '1'
                },
                '0xdac17f958d2ee523a2206206994597c13d831ec7': {
                    'id': '0xdac17f958d2ee523a2206206994597c13d831ec7',
                    'tag': 'ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7',
                    'symbol': 'USDT',
                    'decimals': 6,
                    'chainType': 'ethereum',
                    'chainID': '1'
                },
                '0x72247989079da354c9f0a6886b965bcc86550f8a': {
                    'id': '0x72247989079da354c9f0a6886b965bcc86550f8a',
                    'tag': 'everpay-acnh-0x72247989079da354c9f0a6886b965bcc86550f8a',
                    'symbol': 'ACNH',
                    'decimals': 8,
                    'chainType': 'everpay',
                    'chainID': '1'
                }
            }
        }
    },
    'stakePools': ['basic'],
    'tokenSymbol': 'halo',
    'tokenTotalSupply': "1000000000000000000000000000",
    'tokenDecimals': 18,
    'tokenBalance':{
        'incentive': "500000000000000000000000000", 
        'ecosystem': "190000001920000000000000000",
        'investor': "150000000000000000000000000",
        'team': "150000000000000000000000000",
    },
    'tokenStake':{
        "0xba86Eb14e0fE25caBef156604eBA6D09b226BE6E": {
            "basic": "603448160000000000000000"
        },
        "-tbRiW2AgAZ4W4O3WhV6IFxBgioPK-EHQat1-U0jWgk": {
            "basic": "431034400000000000000000"
        },
        "0xA640C98260F779b6b2E7613218Dd20fEfbdBa957": {
            "basic": "431034400000000000000000"
        },
        "0x3Cc477d246CaFcBe95Df955973b017e29567964F": {
            "basic": "431034400000000000000000"
        },
        "cJZNo30TJPhY4zXLK9IzZaXgN2eKOgbMe0Jg9eVpxfE": {
            "basic": "258620640000000000000000"
        },
        "0x000004302F29bb08B3159EE0c8690f3e52A072e2": {
            "basic": "258620640000000000000000"
        },
        "0xC21a0C238B9bee3BECd5e70c689f23fC5e3F1199": {
            "basic": "258620640000000000000000"
        },
        "0xbf85870b162B2A25a17d68e66eaC31a82F459B36": {
            "basic": "258620640000000000000000"
        },
        "0xd51545B3D65D67B5Cc6aa5CaFb7B45dd296b24aB": {
            "basic": "258620640000000000000000"
        },
        "0x68Af77A050054d4A4E57Ddb257939a5E536F0518": {
            "basic": "258620640000000000000000"
        },
        "T0FOWimx4073_gV2hA-EKht41t5BDEmxhqglC3DH1ns": {
            "basic": "172413760000000000000000"
        },
        "0x57b794f29e32Ec868Ca4758245Ba8c7eF1d143E7": {
            "basic": "172413760000000000000000"
        },
        "0x71fd9230395e58d49D0E4b08F87524D5d004f4C3": {
            "basic": "172413760000000000000000"
        },
        "0x258D9aeF9184d9e21f9b882E243c42fAC1466A59": {
            "basic": "172413760000000000000000"
        },
        "0xb6E047c0fD1d19dDd25023153EbD6bc6fF6Ba9A7": {
            "basic": "172413760000000000000000"
        },
        "0xbf694EBb561529CC36bbf6313a58Ddf8C9cD553B": {
            "basic": "172413760000000000000000"
        },
        "0x3EFBC7180FcA68e2F4Db1AE067a4BA402960c12c": {
            "basic": "172413760000000000000000"
        },
        "0xa2026731B31E4DFBa78314bDBfBFDC8cF5F761F8": {
            "basic": "172413760000000000000000"
        },
        "0x48cC5c98840d7C9291A245f5d8d0585aB3079730": {
            "basic": "172413760000000000000000"
        },
        "0x7966A9951771d982fA3Fa99bCb529F9D387Cfeca": {
            "basic": "172413760000000000000000"
        },
        "0x7346848c1a5b643912D9593E2c9dd051c015FCB3": {
            "basic": "172413760000000000000000"
        },
        "0xC65d7a63212b331DBFEEec55726E2dbd9F54cDD5": {
            "basic": "172413760000000000000000"
        },
        "0xDc19464589c1cfdD10AEdcC1d09336622b282652": {
            "basic": "172413760000000000000000"
        },
        "0xE073715ab98fB49F700fBDbA0FC53C61E9DEA028": {
            "basic": "172413760000000000000000"
        },
        "0xe1702a612C954273A9caD37E7Bc68f1Eccf3Ad94": {
            "basic": "172413760000000000000000"
        },
        "0x48eC8cDbd6b6e588D74631dD8A8d2a8C03cb34Fc": {
            "basic": "172413760000000000000000"
        },
        "hctnleYoGzMf55h3C4pnG3luEUElXkvJj7XWSv7NpUM": {
            "basic": "86206880000000000000000"
        },
        "9eJqlQlOr-6AHqevqIFN-lAZ3AFE0_AKkjQa6qCKP1I": {
            "basic": "86206880000000000000000"
        },
        "AUM4jJULbwYEhsL2EJ46ozRCbEnNy79IE-s4mSD7t-c": {
            "basic": "86206880000000000000000"
        },
        "yyISnAMZUBuJ1bRxtDWAPUl3irF_5yP6V6LSjCInIsY": {
            "basic": "86206880000000000000000"
        },
        "cSYOy8-p1QFenktkDBFyRM3cwZSTrQ_J4EsELLho_UE": {
            "basic": "86206880000000000000000"
        },
        "vu3GICa6cX2SJsKCG-1wtUOCtTqRvRetVXSwoD8cykQ": {
            "basic": "86206880000000000000000"
        },
        "1jowGhPLeXCN1j_OJl0JNqSsQcmNrMkt0SpoZm8S6ts": {
            "basic": "86206880000000000000000"
        },
        "v_k7RH2lUkybFQTMV14D1VNUNUKfw01VSMj_347thtU": {
            "basic": "86206880000000000000000"
        },
        "0x16a56F3a24E467E18Dfb18D87A81B2c588f98f0A": {
            "basic": "86206880000000000000000"
        },
        "0xC57b522Ef07F85E994Ccc17ab4e8d4EDA5f9B54C": {
            "basic": "86206880000000000000000"
        },
        "0x4839460F05b2a8fC1b615e0523EbC46eFba75420": {
            "basic": "86206880000000000000000"
        },
        "0x426fDc4B4F98988326f3c0d7EDf462F5bD55258a": {
            "basic": "86206880000000000000000"
        },
        "0x4680794c6d66ae0bE94cBF0d18a998743c09237e": {
            "basic": "86206880000000000000000"
        },
        "0x0333c37CD77148D71F08DdA7307BBAdc465bE7b1": {
            "basic": "86206880000000000000000"
        },
        "0xc24e9E7983d158c6A715262aa1662F4683A4Ae84": {
            "basic": "86206880000000000000000"
        },
        "0x378cF29359F72862A78995e665eC305763003B65": {
            "basic": "86206880000000000000000"
        },
        "0xCeFA845FC6C18f28F2D1d18D8D85A8ef1E003B0D": {
            "basic": "86206880000000000000000"
        },
        "0xE618FCbC9e2F672E0c20030CF8459c115Da4B931": {
            "basic": "86206880000000000000000"
        },
        "0x601298B7385a855803e1C7f88e89801B5dcfC796": {
            "basic": "86206880000000000000000"
        },
        "0x410ca4870c86af1BC8ae0bab531822E84b63794f": {
            "basic": "86206880000000000000000"
        },
        "0x1F6AB0f279477B0c23CCc824dcB126143Ab6Bd0E": {
            "basic": "86206880000000000000000"
        },
        "0x7C3d60B9e092796D0961989C19a8dC6eF5A3efC8": {
            "basic": "86206880000000000000000"
        },
        "0xF672A5a266FE05DFB8d3Fa6cCF548793B5DE76Bf": {
            "basic": "86206880000000000000000"
        },
        "0x5d4fE11CD67cf21a5B7e75efb12388cD2C250B5a": {
            "basic": "86206880000000000000000"
        },
        "0xa7cD40bD909111e21bC6ed51D7cA98Ba7665Ec73": {
            "basic": "86206880000000000000000"
        },
        "0xc0E0422155b416a53Cdd09E594bDDEDa56e3F952": {
            "basic": "86206880000000000000000"
        },
        "0x1005Ca53422B2b9a73fdfEeB46283B82A36CE215": {
            "basic": "86206880000000000000000"
        },
        "0x7ae421e2d5458FA9FaC04b73f086Bf8dE949F3Ab": {
            "basic": "86206880000000000000000"
        },
        "0x2fAb5BC49b8140a4D601E0E906207b9ee6CEa29A": {
            "basic": "86206880000000000000000"
        },
        "0xa2EB99950E1ee700f0C48644387b8CA460e3EE39": {
            "basic": "86206880000000000000000"
        },
        "0x963Cbd84D9281563A3455cF9f0cd34B6342f0db8": {
            "basic": "86206880000000000000000"
        },
        "0xB6D6d6e5591e1d8A2b64e7362c0FB2eD337d4972": {
            "basic": "86206880000000000000000"
        },
        "0xf71946496600e1e1d47b8A77EB2f109Fd82dc86a": {
            "basic": "86206880000000000000000"
        },
        "0x49b1c6480BA1E332BEBd03e749C0De544FB46BE7": {
            "basic": "86206880000000000000000"
        },
        "0x573394b77fC17F91E9E67F147A9ECe24d67C5073": {
            "basic": "86206880000000000000000"
        },
        "0x39FBD0F137740F9F2878048941d3f4737fD197a3": {
            "basic": "86206880000000000000000"
        },
        "0x1263d65Ca52bbC5CD8068E030a85c90b5339cAcC": {
            "basic": "86206880000000000000000"
        },
        "0x9B546bF012209F5f6bD2fdD0E69e27909cbe5e38": {
            "basic": "86206880000000000000000"
        },
        "0x52c433B5984B2F9ccbA93070Df798A250c2Ea9e4": {
            "basic": "86206880000000000000000"
        },
        "0xE3400Cab2F8d95614704bF68CA6964098c75c486": {
            "basic": "86206880000000000000000"
        },
        "0x4070Caf1CEA63fc038122Bbc9F2e0195387Ae031": {
            "basic": "86206880000000000000000"
        },
        "0xe3F87C2cEE70989eb42B604B5ec9e35C0578E51c": {
            "basic": "86206880000000000000000"
        },
        "0xf2C2b4e87E4c079C5b6edb5a54d3b4092FFb5464": {
            "basic": "86206880000000000000000"
        },
        "0xdAD73954A27E497b9D8eEF95EaE963f130184e33": {
            "basic": "86206880000000000000000"
        }
    }
}

acc = everpay.Account(api_server, signer)
t, result = acc.transfer('eth', halo_address, 0, json.dumps(genesis_params))
print(t.ever_hash, result)