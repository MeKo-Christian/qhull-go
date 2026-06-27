/* Dump Qhull's projected points + key globals after qh_new_qhull, to ground the
   Go port numerically. Input: n, then n lines "x y" (already RAW, tool subtracts mean). */
#include "libqhull_r/qhull_ra.h"
#include <stdio.h>
#include <stdlib.h>
int main(void){
  int n; if(scanf("%d",&n)!=1) return 1;
  double *x=malloc(n*sizeof(double)),*y=malloc(n*sizeof(double));
  for(int i=0;i<n;i++) if(scanf("%lf %lf",&x[i],&y[i])!=2) return 1;
  double xm=0,ym=0; for(int i=0;i<n;i++){xm+=x[i];ym+=y[i];} xm/=n; ym/=n;
  coordT *pts=malloc(n*2*sizeof(coordT));
  for(int i=0;i<n;i++){pts[2*i]=x[i]-xm;pts[2*i+1]=y[i]-ym;}
  qhT qh_qh; qhT *qh=&qh_qh; FILE *ef=fopen("/dev/null","w"); qh_zero(qh,ef);
  int code=qh_new_qhull(qh,2,n,pts,0,(char*)"qhull d Qt Qbb Qc Qz",NULL,ef);
  if(code){printf("ERR %d\n",code);return 1;}
  printf("hull_dim=%d num_points=%d ATinfinity=%d SCALElast=%d NARROWhull=%d\n",qh->hull_dim,qh->num_points,qh->ATinfinity,qh->SCALElast,qh->NARROWhull);
  printf("FLAGS MERGING=%d BESToutside=%d KEEPcoplanar=%d PREmerge=%d MERGEexact=%d ONLYgood=%d UPPERdelaunay=%d DELAUNAY=%d\n",qh->MERGING,qh->BESToutside,qh->KEEPcoplanar,qh->PREmerge,qh->MERGEexact,qh->ONLYgood,qh->UPPERdelaunay,qh->DELAUNAY);
  printf("DISToutside=%.17g MAXcoplanar=%.17g WIDEfacet=%.17g ZEROdelaunay-ish ANGLEround*ZEROdel=%.17g\n",qh_DISToutside,qh->MAXcoplanar,qh->WIDEfacet,qh->ANGLEround);
  printf("DISTround=%.17g ANGLEround=%.17g MAXabs_coord=%.17g MAXwidth=%.17g MAXsumcoord=%.17g\n",
         qh->DISTround,qh->ANGLEround,qh->MAXabs_coord,qh->MAXwidth,qh->MAXsumcoord);
  printf("last_low=%.17g last_high=%.17g last_newhigh=%.17g MINlast=%.17g MAXlast=%.17g\n",
         qh->last_low,qh->last_high,qh->last_newhigh,qh->MINlastcoord,qh->MAXlastcoord);
  printf("MINoutside=%.17g MINvisible=%.17g min_vertex=%.17g\n",qh->MINoutside,qh->MINvisible,qh->min_vertex);
  printf("POINTS (hull_dim*num_points, includes infinity last):\n");
  for(int i=0;i<qh->num_points;i++){
    printf("  p%d:",i);
    for(int k=0;k<qh->hull_dim;k++) printf(" %.17g",qh->first_point[i*qh->hull_dim+k]);
    printf("\n");
  }
  return 0;
}
