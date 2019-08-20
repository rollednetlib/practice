#include <libnet.h>

int main(int argc, char *argv[])
{
	if ( argv[1] == NULL ) 
	{
		printf("\n	Usage: %s [nic]\n\n",argv[0]);
		return 2;
	}


	char err_buf[LIBNET_ERRBUF_SIZE];
	libnet_t *l;
	l=libnet_init(LIBNET_RAW4, argv[1], err_buf);
	if ( l == NULL ) { printf("\nFailed to init\n\n"); return 2; }

	// Define target IP //
	char ip_addr_str[16]; 
	printf("\nEnter Target IP: ");
	scanf("%15s",ip_addr_str);

	uint32_t ip_addr;
	ip_addr = libnet_name2addr4(l,ip_addr_str,LIBNET_DONT_RESOLVE);

	uint8_t *addp;
	addp = (uint8_t*)(&ip_addr);

	// Build icmp header//
	uint8_t type, code;
	uint16_t id, seq, sum;
	char payload[16];

	printf("Enter data to roll!\n> ");
	scanf("%s",payload);
	libnet_build_icmpv4_echo(ICMP_ECHO,0,0,id,seq,(uint8_t*)payload,sizeof(payload),l,0);

	libnet_autobuild_ipv4(LIBNET_IPV4_H + LIBNET_ICMPV4_ECHO_H + sizeof(payload), IPPROTO_ICMP, ip_addr, l);

	int bytes_written;
	bytes_written = libnet_write(l);
	printf("%d WRITE!!\n\n",bytes_written);
	libnet_destroy(l);
	
	return 0;
}
