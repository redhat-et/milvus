variable "vpc_id" {
  type = string
}

variable "ssh_public_key_path" {
  type    = string
  default = "~/.ssh/id_rsa.pub"
}

variable "ssh_private_key_path" {
  type    = string
  default = "~/.ssh/id_rsa"
}


variable "ami_id" {
  type    = string
  default = "ami-0d77c9d87c7e619f9"
}

variable "rh_access" {
  sensitive = true
  type      = string
}

variable "rh_org" {
  sensitive = true
  type      = string
}


// generate a new security group to allow ssh and https traffic
resource "aws_security_group" "builder-access" {
  name        = "builder-access"
  description = "Allow ssh and https traffic"
  vpc_id      = var.vpc_id

  ingress {
    description = "SSH from VPC"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_key_pair" "sigkey" {
  key_name   = uuid()
  public_key = file(var.ssh_public_key_path)
}

resource "aws_instance" "builder" {
  ami                    = var.ami_id
  instance_type          = "m5.large"
  associate_public_ip_address = true
  vpc_security_group_ids = [aws_security_group.builder-access.id]
  key_name               = aws_key_pair.sigkey.key_name

  provisioner "remote-exec" {
    inline = [
      "sudo cloud-init status --wait",
      "echo 'Connection Established'",
      "sudo subscription-manager register --activationkey=${var.rh_access} --org=${var.rh_org} --force",
      "sudo dnf -y install container-tools podman",
      "sudo subscription-manager config --rhsm.manage_repos=1",
      "systemctl enable --now podman.socket --user",
    ]
  }
  connection {
    type        = "ssh"
    user        = "ec2-user"
    host        = self.public_ip
    private_key = file(var.ssh_private_key_path)
  }
}

// Output public ip address
output "public_ip" {
  value = aws_instance.builder.public_ip
}
