#include <stdio.h>
#include <string.h>
void main(){
	struct bookdetails{
		char name[20];
		char author[20];
		int id;
		float price;
	};
	int i,num,a;
	int bookid;
	char bname[20];
	struct bookdetails b[10];
	do{
		printf("\n1.add new books");
		printf("\n2.search a book by id");
		printf("\n3.search a book by name");
		printf("\n4.display books");
		printf("\n5.exit");
		printf("\n\n\nsearch operation : ");
		
		scanf("%d",&a);
		switch(a){
			case 1:
                printf("enter the number of books to be entered : ");
	            scanf("%d",&num);
	            for(i=0;i<num;i++){
		            printf("\n book %d details",i+1);
		
		            printf("\nenter the book name : ");
		            scanf("%s",b[i].name);
		            printf("\nenter the author name : ");
		            scanf("%s",b[i].author);
		            printf("\nenter the book id : ");
		            scanf("%d",&b[i].id);
		            printf("\nenter the book price : ");
		            scanf("%f",&b[i].price);
	            }
			    break;
			case 2:
			
				printf("enter the book id to be search : ");
				scanf("%d",&bookid);
				for(i=0;i<num;i++){
				    if(bookid==b[i].id);
				        break;
				}
				if(i<num){
				    printf("\n");
		            printf("book number %d\n",i);
		            printf("\tbook name is=%s \n",b[i].name);
		            printf("\tbook author is=%s \n",b[i].author);
		            printf("\tbook pages is=%d \n",b[i].id);
		            printf("\tbook name is=%f \n",b[i].price);
		            printf("\n"); 
				}
				else
				    printf("book not found");
			    break;
			case 3:
				while(b[num].name=='\0'){
					num++;
				}
				printf("enter the book name to be search");
				scanf("%s",bname);
			    for(i=0;i<num;i++){
			    	if(strcmp(b[i].name,bname)==0){
			    		break;
					}
				}
				if(i<num){
				    printf("\n");
		            printf("book number %d\n",i);
		            printf("\tbook name is=%s \n",b[i].name);
		            printf("\tbook author is=%s \n",b[i].author);
		            printf("\tbook pages is=%d \n",b[i].id);
		            printf("\tbook name is=%f \n",b[i].price);
		            printf("\n"); 
				}
				else
				    printf("book not found");
			    break;
			case 4:
				while(b[num].name=='\0'){
					num++;
				}
				int t=1;
            	for(i=0;i<num;i++,t++){
		        printf("\n");
		        printf("book number %d\n",t);
		        printf("\tbook %d name is=%s \n",t,b[i].name);
		        printf("\tbook %d author is=%s \n",t,b[i].author);
		        printf("\tbook %d pages is=%d \n",t,b[i].id);
		        printf("\tbook %d name is=%f \n",t,b[i].price);
	         	printf("\n");
	            }
	            break;
	        case 5:
	        	printf("\n program exit.....");
	        	break;
				
		}
		
	}while (a!=5);
	return 0;
}
