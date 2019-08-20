#include <stdio.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>

/*
 * A simple program to sub for sockets
 *
 */


int sw_listen(int argc, char *argv[])
{
	int port = atoi(argv[2]);
	printf("Listening on port: %d\n",port);
	
	int sfdl = socket(AF_INET, SOCK_STREAM, 0);

	struct sockaddr_in sockl;
	sockl.sin_family = AF_INET;
	sockl.sin_port = htons(atoi(argv[2]));
	sockl.sin_addr.s_addr = INADDR_ANY;

	int statusl = bind( sfdl,(struct sockaddr *) &sockl, sizeof(sockl));
	if ( statusl != 0 )
	{
		perror("ERROR: ");
		return 2;
	}
	
	listen(sfdl,5);
	
	/*int client_sock = accept(sfd, NULL, NULL);
	(client_sock, */

	int client_sock = accept(sfdl, NULL, NULL);

	char bufferl[256];
	recv(client_sock, &bufferl, sizeof(bufferl), 0);

	printf("\n%s\n",bufferl);

	close(sfdl);
	return 0;

}

int sw_connect(int argc, char *argv[])
{
	int port = atoi(argv[2]);
	/* socket init */
	int sfd = socket(AF_INET, SOCK_STREAM, 0);
	struct sockaddr_in sock;
	sock.sin_family = AF_INET;
	sock.sin_port = htons(atoi(argv[2]));
	inet_aton(argv[1], &sock.sin_addr.s_addr);

	int status = connect(sfd, (struct sockaddr *) &sock, sizeof(sock));
	if ( status != 0 )
	{
		perror("FATAL ERROR: ");
		return 2;
	}

	char buffer[256];
	recv(sfd, &buffer, sizeof(buffer), 0);

	printf("\n%s\n",buffer);

	close(sfd);
	return 0;
}


int main(int argc, char *argv[])
{
	if ( argc < 2 )
	{ 
		printf("\n\tUsage STATEMENT\n\n\n\tConnect:(waits on recv)\n\tmal 192.168.1.1 22\n\n\tListen: (Waits on recv)\n\tmal 222\n\n");
		return 2;
	} else if ( strcmp(argv[1], "-l") == 0 )
	{
		sw_listen(argc, argv);
	} else {
		sw_connect(argc, argv);
	}
	return 0;
}
