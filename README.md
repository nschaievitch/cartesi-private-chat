# cartesi-private-chat

This project uses Cartesi to build a fully decentralized protocol for conference key generation using the Burmester-Desmedt protocol. This allows for a single secret (password) to be shared by all members of a group, and this can be used to create private messaging without the need to encrypt messaging for each member of the group. 

The project also includes a **State transition** database for each group. This allows each group to hold a separate encrypted state machine to be run by each member, which means it allows for arbitrary **fully private** decentralized applications, without adding much extra computation (constant time).

Some uses of this, for example, could be:
- Private federated social media
- Private federated chat rooms
- Private federated games

And much more.

The current client implements private chat rooms. It uses a custom implementation of the cryptographic primitives running fully on the browser with a web-assembly implementation

## Instructions 
Make sure you have the following dependencies installed:
- **sunodo**
- **yarn**

### Steps (running)
1. Run the `install-libraries.sh` script to install client and backend libraries.
2. On separate threads (separate terminals for convenience), run the `start-backend.sh`, `start-cartesi.sh`, and `start-frontend.sh` scripts.

### Steps (usage)
1. On the client, select a wallet to use. The example client includes fake wallets for convinience, select (0-9). You can use a browser wallet by selecting -1.
2. Enter the addresses of the group, one by line, into the text area and press create group.
3. Once the group is created, it will appear on the table below.
4. All members must first `join` the group, and then `sign`.
5. Once all members have joined and signed into the group, press `visit` to enter the chat room. 